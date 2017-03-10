// mgo - MongoDB driver for Go
//
// Copyright (c) 2010-2012 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Mode int

const (
	// Relevant documentation on read preference modes:
	//
	//     http://docs.mongodb.org/manual/reference/read-preference/
	//
	Primary            Mode = 2 // Default mode. All operations read from the current replica set primary.
	PrimaryPreferred   Mode = 3 // Read from the primary if available. Read from the secondary otherwise.
	Secondary          Mode = 4 // Read from one of the nearest secondary members of the replica set.
	SecondaryPreferred Mode = 5 // Read from one of the nearest secondaries if available. Read from primary otherwise.
	Nearest            Mode = 6 // Read from one of the nearest members, irrespective of it being primary or secondary.

	// Read preference modes are specific to mgo:
	Eventual  Mode = 0 // Same as Nearest, but may change servers between reads.
	Monotonic Mode = 1 // Same as SecondaryPreferred before first write. Same as Primary after first write.
	Strong    Mode = 2 // Same as Primary.
)

// mgo.v3: Drop Strong mode, suffix all modes with "Mode".

// When changing the Session type, check if newSession and copySession
// need to be updated too.

// Session represents a communication session with the database.
//
// All Session methods are concurrency-safe and may be called from multiple
// goroutines. In all session modes but Eventual, using the session from
// multiple goroutines will cause them to share the same underlying socket.
// See the documentation on Session.SetMode for more details.
type Session struct {
	m                sync.RWMutex
	cluster_         *mongoCluster
	slaveSocket      *mongoSocket
	masterSocket     *mongoSocket
	slaveOk          bool
	consistency      Mode
	queryConfig      query
	safeOp           *queryOp
	syncTimeout      time.Duration
	sockTimeout      time.Duration
	defaultdb        string
	sourcedb         string
	dialCred         *Credential
	creds            []Credential
	poolLimit        int
	bypassValidation bool
}

type Database struct {
	Session *Session
	Name    string
}

type Collection struct {
	Database *Database
	Name     string // "collection"
	FullName string // "db.collection"
}

type Query struct {
	m       sync.Mutex
	session *Session
	query   // Enables default settings in session.
}

type query struct {
	op       queryOp
	prefetch float64
	limit    int32
}

type getLastError struct {
	CmdName  int         "getLastError,omitempty"
	W        interface{} "w,omitempty"
	WTimeout int         "wtimeout,omitempty"
	FSync    bool        "fsync,omitempty"
	J        bool        "j,omitempty"
}

type Iter struct {
	m              sync.Mutex
	gotReply       sync.Cond
	session        *Session
	server         *mongoServer
	docData        queue
	err            error
	op             getMoreOp
	prefetch       float64
	limit          int32
	docsToReceive  int
	docsBeforeMore int
	timeout        time.Duration
	timedout       bool
	findCmd        bool
}

var (
	ErrNotFound = errors.New("not found")
	ErrCursor   = errors.New("invalid cursor")
)

const defaultPrefetch = 0.25

// Dial establishes a new session to the cluster identified by the given seed
// server(s). The session will enable communication with all of the servers in
// the cluster, so the seed servers are used only to find out about the cluster
// topology.
//
// Dial will timeout after 10 seconds if a server isn't reached. The returned
// session will timeout operations after one minute by default if servers
// aren't available. To customize the timeout, see DialWithTimeout,
// SetSyncTimeout, and SetSocketTimeout.
//
// This method is generally called just once for a given cluster.  Further
// sessions to the same cluster are then established using the New or Copy
// methods on the obtained session. This will make them share the underlying
// cluster, and manage the pool of connections appropriately.
//
// Once the session is not useful anymore, Close must be called to release the
// resources appropriately.
//
// The seed servers must be provided in the following format:
//
//     [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
//
// For example, it may be as simple as:
//
//     localhost
//
// Or more involved like:
//
//     mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
//
// If the port number is not provided for a server, it defaults to 27017.
//
// The username and password provided in the URL will be used to authenticate
// into the database named after the slash at the end of the host names, or
// into the "admin" database if none is provided.  The authentication information
// will persist in sessions obtained through the New method as well.
//
// The following connection options are supported after the question mark:
//
//     connect=direct
//
//         Disables the automatic replica set server discovery logic, and
//         forces the use of servers provided only (even if secondaries).
//         Note that to talk to a secondary the consistency requirements
//         must be relaxed to Monotonic or Eventual via SetMode.
//
//
//     connect=replicaSet
//
//  	   Discover replica sets automatically. Default connection behavior.
//
//
//     replicaSet=<setname>
//
//         If specified will prevent the obtained session from communicating
//         with any server which is not part of a replica set with the given name.
//         The default is to communicate with any server specified or discovered
//         via the servers contacted.
//
//
//     authSource=<db>
//
//         Informs the database used to establish credentials and privileges
//         with a MongoDB server. Defaults to the database name provided via
//         the URL path, and "admin" if that's unset.
//
//
//     authMechanism=<mechanism>
//
//        Defines the protocol for credential negotiation. Defaults to "MONGODB-CR",
//        which is the default username/password challenge-response mechanism.
//
//
//     gssapiServiceName=<name>
//
//        Defines the service name to use when authenticating with the GSSAPI
//        mechanism. Defaults to "mongodb".
//
//
//     maxPoolSize=<limit>
//
//        Defines the per-server socket pool limit. Defaults to 4096.
//        See Session.SetPoolLimit for details.
//
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/reference/connection-string/
//
func Dial(url string) (*Session, error) {
	session, err := DialWithTimeout(url, 10*time.Second)
	if err == nil {
		session.SetSyncTimeout(1 * time.Minute)
		session.SetSocketTimeout(1 * time.Minute)
	}
	return session, err
}

// DialWithTimeout works like Dial, but uses timeout as the amount of time to
// wait for a server to respond when first connecting and also on follow up
// operations in the session. If timeout is zero, the call may block
// forever waiting for a connection to be made.
//
// See SetSyncTimeout for customizing the timeout for the session.
func DialWithTimeout(url string, timeout time.Duration) (*Session, error) {
	info, err := ParseURL(url)
	if err != nil {
		return nil, err
	}
	info.Timeout = timeout
	return DialWithInfo(info)
}

// ParseURL parses a MongoDB URL as accepted by the Dial function and returns
// a value suitable for providing into DialWithInfo.
//
// See Dial for more details on the format of url.
func ParseURL(url string) (*DialInfo, error) {
	uinfo, err := extractURL(url)
	if err != nil {
		return nil, err
	}
	direct := false
	mechanism := ""
	service := ""
	source := ""
	setName := ""
	poolLimit := 0
	for k, v := range uinfo.options {
		switch k {
		case "authSource":
			source = v
		case "authMechanism":
			mechanism = v
		case "gssapiServiceName":
			service = v
		case "replicaSet":
			setName = v
		case "maxPoolSize":
			poolLimit, err = strconv.Atoi(v)
			if err != nil {
				return nil, errors.New("bad value for maxPoolSize: " + v)
			}
		case "connect":
			if v == "direct" {
				direct = true
				break
			}
			if v == "replicaSet" {
				break
			}
			fallthrough
		default:
			return nil, errors.New("unsupported connection URL option: " + k + "=" + v)
		}
	}
	info := DialInfo{
		Addrs:          uinfo.addrs,
		Direct:         direct,
		Database:       uinfo.db,
		Username:       uinfo.user,
		Password:       uinfo.pass,
		Mechanism:      mechanism,
		Service:        service,
		Source:         source,
		PoolLimit:      poolLimit,
		ReplicaSetName: setName,
	}
	return &info, nil
}

// DialInfo holds options for establishing a session with a MongoDB cluster.
// To use a URL, see the Dial function.
type DialInfo struct {
	// Addrs holds the addresses for the seed servers.
	Addrs []string

	// Direct informs whether to establish connections only with the
	// specified seed servers, or to obtain information for the whole
	// cluster and establish connections with further servers too.
	Direct bool

	// Timeout is the amount of time to wait for a server to respond when
	// first connecting and on follow up operations in the session. If
	// timeout is zero, the call may block forever waiting for a connection
	// to be established. Timeout does not affect logic in DialServer.
	Timeout time.Duration

	// FailFast will cause connection and query attempts to fail faster when
	// the server is unavailable, instead of retrying until the configured
	// timeout period. Note that an unavailable server may silently drop
	// packets instead of rejecting them, in which case it's impossible to
	// distinguish it from a slow server, so the timeout stays relevant.
	FailFast bool

	// Database is the default database name used when the Session.DB method
	// is called with an empty name, and is also used during the initial
	// authentication if Source is unset.
	Database string

	// ReplicaSetName, if specified, will prevent the obtained session from
	// communicating with any server which is not part of a replica set
	// with the given name. The default is to communicate with any server
	// specified or discovered via the servers contacted.
	ReplicaSetName string

	// Source is the database used to establish credentials and privileges
	// with a MongoDB server. Defaults to the value of Database, if that is
	// set, or "admin" otherwise.
	Source string

	// Service defines the service name to use when authenticating with the GSSAPI
	// mechanism. Defaults to "mongodb".
	Service string

	// ServiceHost defines which hostname to use when authenticating
	// with the GSSAPI mechanism. If not specified, defaults to the MongoDB
	// server's address.
	ServiceHost string

	// Mechanism defines the protocol for credential negotiation.
	// Defaults to "MONGODB-CR".
	Mechanism string

	// Username and Password inform the credentials for the initial authentication
	// done on the database defined by the Source field. See Session.Login.
	Username string
	Password string

	// PoolLimit defines the per-server socket pool limit. Defaults to 4096.
	// See Session.SetPoolLimit for details.
	PoolLimit int

	// DialServer optionally specifies the dial function for establishing
	// connections with the MongoDB servers.
	DialServer func(addr *ServerAddr) (net.Conn, error)

	// WARNING: This field is obsolete. See DialServer above.
	Dial func(addr net.Addr) (net.Conn, error)
}

// mgo.v3: Drop DialInfo.Dial.

// ServerAddr represents the address for establishing a connection to an
// individual MongoDB server.
type ServerAddr struct {
	str string
	tcp *net.TCPAddr
}

// String returns the address that was provided for the server before resolution.
func (addr *ServerAddr) String() string {
	return addr.str
}

// TCPAddr returns the resolved TCP address for the server.
func (addr *ServerAddr) TCPAddr() *net.TCPAddr {
	return addr.tcp
}

// DialWithInfo establishes a new session to the cluster identified by info.
func DialWithInfo(info *DialInfo) (*Session, error) {
	addrs := make([]string, len(info.Addrs))
	for i, addr := range info.Addrs {
		p := strings.LastIndexAny(addr, "]:")
		if p == -1 || addr[p] != ':' {
			// XXX This is untested. The test suite doesn't use the standard port.
			addr += ":27017"
		}
		addrs[i] = addr
	}
	cluster := newCluster(addrs, info.Direct, info.FailFast, dialer{info.Dial, info.DialServer}, info.ReplicaSetName)
	session := newSession(Eventual, cluster, info.Timeout)
	session.defaultdb = info.Database
	if session.defaultdb == "" {
		session.defaultdb = "test"
	}
	session.sourcedb = info.Source
	if session.sourcedb == "" {
		session.sourcedb = info.Database
		if session.sourcedb == "" {
			session.sourcedb = "admin"
		}
	}
	if info.Username != "" {
		source := session.sourcedb
		if info.Source == "" &&
			(info.Mechanism == "GSSAPI" || info.Mechanism == "PLAIN" || info.Mechanism == "MONGODB-X509") {
			source = "$external"
		}
		session.dialCred = &Credential{
			Username:    info.Username,
			Password:    info.Password,
			Mechanism:   info.Mechanism,
			Service:     info.Service,
			ServiceHost: info.ServiceHost,
			Source:      source,
		}
		session.creds = []Credential{*session.dialCred}
	}
	if info.PoolLimit > 0 {
		session.poolLimit = info.PoolLimit
	}
	cluster.Release()

	// People get confused when we return a session that is not actually
	// established to any servers yet (e.g. what if url was wrong). So,
	// ping the server to ensure there's someone there, and abort if it
	// fails.
	if err := session.Ping(); err != nil {
		session.Close()
		return nil, err
	}
	session.SetMode(Strong, true)
	return session, nil
}

func isOptSep(c rune) bool {
	return c == ';' || c == '&'
}

type urlInfo struct {
	addrs   []string
	user    string
	pass    string
	db      string
	options map[string]string
}

func extractURL(s string) (*urlInfo, error) {
	if strings.HasPrefix(s, "mongodb://") {
		s = s[10:]
	}
	info := &urlInfo{options: make(map[string]string)}
	if c := strings.Index(s, "?"); c != -1 {
		for _, pair := range strings.FieldsFunc(s[c+1:], isOptSep) {
			l := strings.SplitN(pair, "=", 2)
			if len(l) != 2 || l[0] == "" || l[1] == "" {
				return nil, errors.New("connection option must be key=value: " + pair)
			}
			info.options[l[0]] = l[1]
		}
		s = s[:c]
	}
	if c := strings.Index(s, "@"); c != -1 {
		pair := strings.SplitN(s[:c], ":", 2)
		if len(pair) > 2 || pair[0] == "" {
			return nil, errors.New("credentials must be provided as user:pass@host")
		}
		var err error
		info.user, err = url.QueryUnescape(pair[0])
		if err != nil {
			return nil, fmt.Errorf("cannot unescape username in URL: %q", pair[0])
		}
		if len(pair) > 1 {
			info.pass, err = url.QueryUnescape(pair[1])
			if err != nil {
				return nil, fmt.Errorf("cannot unescape password in URL")
			}
		}
		s = s[c+1:]
	}
	if c := strings.Index(s, "/"); c != -1 {
		info.db = s[c+1:]
		s = s[:c]
	}
	info.addrs = strings.Split(s, ",")
	return info, nil
}

func newSession(consistency Mode, cluster *mongoCluster, timeout time.Duration) (session *Session) {
	cluster.Acquire()
	session = &Session{
		cluster_:    cluster,
		syncTimeout: timeout,
		sockTimeout: timeout,
		poolLimit:   4096,
	}
	debugf("New session %p on cluster %p", session, cluster)
	session.SetMode(consistency, true)
	session.SetSafe(&Safe{})
	session.queryConfig.prefetch = defaultPrefetch
	return session
}

func copySession(session *Session, keepCreds bool) (s *Session) {
	cluster := session.cluster()
	cluster.Acquire()
	if session.masterSocket != nil {
		session.masterSocket.Acquire()
	}
	if session.slaveSocket != nil {
		session.slaveSocket.Acquire()
	}
	var creds []Credential
	if keepCreds {
		creds = make([]Credential, len(session.creds))
		copy(creds, session.creds)
	} else if session.dialCred != nil {
		creds = []Credential{*session.dialCred}
	}
	scopy := *session
	scopy.m = sync.RWMutex{}
	scopy.creds = creds
	s = &scopy
	debugf("New session %p on cluster %p (copy from %p)", s, cluster, session)
	return s
}

// LiveServers returns a list of server addresses which are
// currently known to be alive.
func (s *Session) LiveServers() (addrs []string) {
	s.m.RLock()
	addrs = s.cluster().LiveServers()
	s.m.RUnlock()
	return addrs
}

// DB returns a value representing the named database. If name
// is empty, the database name provided in the dialed URL is
// used instead. If that is also empty, "test" is used as a
// fallback in a way equivalent to the mongo shell.
//
// Creating this value is a very lightweight operation, and
// involves no network communication.
func (s *Session) DB(name string) *Database {
	if name == "" {
		name = s.defaultdb
	}
	return &Database{s, name}
}

// C returns a value representing the named collection.
//
// Creating this value is a very lightweight operation, and
// involves no network communication.
func (db *Database) C(name string) *Collection {
	return &Collection{db, name, db.Name + "." + name}
}

// With returns a copy of db that uses session s.
func (db *Database) With(s *Session) *Database {
	newdb := *db
	newdb.Session = s
	return &newdb
}

// With returns a copy of c that uses session s.
func (c *Collection) With(s *Session) *Collection {
	newdb := *c.Database
	newdb.Session = s
	newc := *c
	newc.Database = &newdb
	return &newc
}

// GridFS returns a GridFS value representing collections in db that
// follow the standard GridFS specification.
// The provided prefix (sometimes known as root) will determine which
// collections to use, and is usually set to "fs" when there is a
// single GridFS in the database.
//
// See the GridFS Create, Open, and OpenId methods for more details.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/GridFS
//     http://www.mongodb.org/display/DOCS/GridFS+Tools
//     http://www.mongodb.org/display/DOCS/GridFS+Specification
//
func (db *Database) GridFS(prefix string) *GridFS {
	return newGridFS(db, prefix)
}

// Run issues the provided command on the db database and unmarshals
// its result in the respective argument. The cmd argument may be either
// a string with the command name itself, in which case an empty document of
// the form bson.M{cmd: 1} will be used, or it may be a full command document.
//
// Note that MongoDB considers the first marshalled key as the command
// name, so when providing a command with options, it's important to
// use an ordering-preserving document, such as a struct value or an
// instance of bson.D.  For instance:
//
//     db.Run(bson.D{{"create", "mycollection"}, {"size", 1024}})
//
// For privilleged commands typically run on the "admin" database, see
// the Run method in the Session type.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Commands
//     http://www.mongodb.org/display/DOCS/List+of+Database+CommandSkips
//
func (db *Database) Run(cmd interface{}, result interface{}) error {
	socket, err := db.Session.acquireSocket(true)
	if err != nil {
		return err
	}
	defer socket.Release()

	// This is an optimized form of db.C("$cmd").Find(cmd).One(result).
	return db.run(socket, cmd, result)
}

// Credential holds details to authenticate with a MongoDB server.
type Credential struct {
	// Username and Password hold the basic details for authentication.
	// Password is optional with some authentication mechanisms.
	Username string
	Password string

	// Source is the database used to establish credentials and privileges
	// with a MongoDB server. Defaults to the default database provided
	// during dial, or "admin" if that was unset.
	Source string

	// Service defines the service name to use when authenticating with the GSSAPI
	// mechanism. Defaults to "mongodb".
	Service string

	// ServiceHost defines which hostname to use when authenticating
	// with the GSSAPI mechanism. If not specified, defaults to the MongoDB
	// server's address.
	ServiceHost string

	// Mechanism defines the protocol for credential negotiation.
	// Defaults to "MONGODB-CR".
	Mechanism string
}

// Login authenticates with MongoDB using the provided credential.  The
// authentication is valid for the whole session and will stay valid until
// Logout is explicitly called for the same database, or the session is
// closed.
func (db *Database) Login(user, pass string) error {
	return db.Session.Login(&Credential{Username: user, Password: pass, Source: db.Name})
}

// Login authenticates with MongoDB using the provided credential.  The
// authentication is valid for the whole session and will stay valid until
// Logout is explicitly called for the same database, or the session is
// closed.
func (s *Session) Login(cred *Credential) error {
	socket, err := s.acquireSocket(true)
	if err != nil {
		return err
	}
	defer socket.Release()

	credCopy := *cred
	if cred.Source == "" {
		if cred.Mechanism == "GSSAPI" {
			credCopy.Source = "$external"
		} else {
			credCopy.Source = s.sourcedb
		}
	}
	err = socket.Login(credCopy)
	if err != nil {
		return err
	}

	s.m.Lock()
	s.creds = append(s.creds, credCopy)
	s.m.Unlock()
	return nil
}

func (s *Session) socketLogin(socket *mongoSocket) error {
	for _, cred := range s.creds {
		if err := socket.Login(cred); err != nil {
			return err
		}
	}
	return nil
}

// Logout removes any established authentication credentials for the database.
func (db *Database) Logout() {
	session := db.Session
	dbname := db.Name
	session.m.Lock()
	found := false
	for i, cred := range session.creds {
		if cred.Source == dbname {
			copy(session.creds[i:], session.creds[i+1:])
			session.creds = session.creds[:len(session.creds)-1]
			found = true
			break
		}
	}
	if found {
		if session.masterSocket != nil {
			session.masterSocket.Logout(dbname)
		}
		if session.slaveSocket != nil {
			session.slaveSocket.Logout(dbname)
		}
	}
	session.m.Unlock()
}

// LogoutAll removes all established authentication credentials for the session.
func (s *Session) LogoutAll() {
	s.m.Lock()
	for _, cred := range s.creds {
		if s.masterSocket != nil {
			s.masterSocket.Logout(cred.Source)
		}
		if s.slaveSocket != nil {
			s.slaveSocket.Logout(cred.Source)
		}
	}
	s.creds = s.creds[0:0]
	s.m.Unlock()
}

// User represents a MongoDB user.
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/reference/privilege-documents/
//     http://docs.mongodb.org/manual/reference/user-privileges/
//
type User struct {
	// Username is how the user identifies itself to the system.
	Username string `bson:"user"`

	// Password is the plaintext password for the user. If set,
	// the UpsertUser method will hash it into PasswordHash and
	// unset it before the user is added to the database.
	Password string `bson:",omitempty"`

	// PasswordHash is the MD5 hash of Username+":mongo:"+Password.
	PasswordHash string `bson:"pwd,omitempty"`

	// CustomData holds arbitrary data admins decide to associate
	// with this user, such as the full name or employee id.
	CustomData interface{} `bson:"customData,omitempty"`

	// Roles indicates the set of roles the user will be provided.
	// See the Role constants.
	Roles []Role `bson:"roles"`

	// OtherDBRoles allows assigning roles in other databases from
	// user documents inserted in the admin database. This field
	// only works in the admin database.
	OtherDBRoles map[string][]Role `bson:"otherDBRoles,omitempty"`

	// UserSource indicates where to look for this user's credentials.
	// It may be set to a database name, or to "$external" for
	// consulting an external resource such as Kerberos. UserSource
	// must not be set if Password or PasswordHash are present.
	//
	// WARNING: This setting was only ever supported in MongoDB 2.4,
	// and is now obsolete.
	UserSource string `bson:"userSource,omitempty"`
}

type Role string

const (
	// Relevant documentation:
	//
	//     http://docs.mongodb.org/manual/reference/user-privileges/
	//
	RoleRoot         Role = "root"
	RoleRead         Role = "read"
	RoleReadAny      Role = "readAnyDatabase"
	RoleReadWrite    Role = "readWrite"
	RoleReadWriteAny Role = "readWriteAnyDatabase"
	RoleDBAdmin      Role = "dbAdmin"
	RoleDBAdminAny   Role = "dbAdminAnyDatabase"
	RoleUserAdmin    Role = "userAdmin"
	RoleUserAdminAny Role = "userAdminAnyDatabase"
	RoleClusterAdmin Role = "clusterAdmin"
)

// UpsertUser updates the authentication credentials and the roles for
// a MongoDB user within the db database. If the named user doesn't exist
// it will be created.
//
// This method should only be used from MongoDB 2.4 and on. For older
// MongoDB releases, use the obsolete AddUser method instead.
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/reference/user-privileges/
//     http://docs.mongodb.org/manual/reference/privilege-documents/
//
func (db *Database) UpsertUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("user has no Username")
	}
	if (user.Password != "" || user.PasswordHash != "") && user.UserSource != "" {
		return fmt.Errorf("user has both Password/PasswordHash and UserSource set")
	}
	if len(user.OtherDBRoles) > 0 && db.Name != "admin" && db.Name != "$external" {
		return fmt.Errorf("user with OtherDBRoles is only supported in the admin or $external databases")
	}

	// Attempt to run this using 2.6+ commands.
	rundb := db
	if user.UserSource != "" {
		// Compatibility logic for the userSource field of MongoDB <= 2.4.X
		rundb = db.Session.DB(user.UserSource)
	}
	err := rundb.runUserCmd("updateUser", user)
	// retry with createUser when isAuthError in order to enable the "localhost exception"
	if isNotFound(err) || isAuthError(err) {
		return rundb.runUserCmd("createUser", user)
	}
	if !isNoCmd(err) {
		return err
	}

	// Command does not exist. Fallback to pre-2.6 behavior.
	var set, unset bson.D
	if user.Password != "" {
		psum := md5.New()
		psum.Write([]byte(user.Username + ":mongo:" + user.Password))
		set = append(set, bson.DocElem{"pwd", hex.EncodeToString(psum.Sum(nil))})
		unset = append(unset, bson.DocElem{"userSource", 1})
	} else if user.PasswordHash != "" {
		set = append(set, bson.DocElem{"pwd", user.PasswordHash})
		unset = append(unset, bson.DocElem{"userSource", 1})
	}
	if user.UserSource != "" {
		set = append(set, bson.DocElem{"userSource", user.UserSource})
		unset = append(unset, bson.DocElem{"pwd", 1})
	}
	if user.Roles != nil || user.OtherDBRoles != nil {
		set = append(set, bson.DocElem{"roles", user.Roles})
		if len(user.OtherDBRoles) > 0 {
			set = append(set, bson.DocElem{"otherDBRoles", user.OtherDBRoles})
		} else {
			unset = append(unset, bson.DocElem{"otherDBRoles", 1})
		}
	}
	users := db.C("system.users")
	err = users.Update(bson.D{{"user", user.Username}}, bson.D{{"$unset", unset}, {"$set", set}})
	if err == ErrNotFound {
		set = append(set, bson.DocElem{"user", user.Username})
		if user.Roles == nil && user.OtherDBRoles == nil {
			// Roles must be sent, as it's the way MongoDB distinguishes
			// old-style documents from new-style documents in pre-2.6.
			set = append(set, bson.DocElem{"roles", user.Roles})
		}
		err = users.Insert(set)
	}
	return err
}

func isNoCmd(err error) bool {
	e, ok := err.(*QueryError)
	return ok && (e.Code == 59 || e.Code == 13390 || strings.HasPrefix(e.Message, "no such cmd:"))
}

func isNotFound(err error) bool {
	e, ok := err.(*QueryError)
	return ok && e.Code == 11
}

func isAuthError(err error) bool {
	e, ok := err.(*QueryError)
	return ok && e.Code == 13
}

func (db *Database) runUserCmd(cmdName string, user *User) error {
	cmd := make(bson.D, 0, 16)
	cmd = append(cmd, bson.DocElem{cmdName, user.Username})
	if user.Password != "" {
		cmd = append(cmd, bson.DocElem{"pwd", user.Password})
	}
	var roles []interface{}
	for _, role := range user.Roles {
		roles = append(roles, role)
	}
	for db, dbroles := range user.OtherDBRoles {
		for _, role := range dbroles {
			roles = append(roles, bson.D{{"role", role}, {"db", db}})
		}
	}
	if roles != nil || user.Roles != nil || cmdName == "createUser" {
		cmd = append(cmd, bson.DocElem{"roles", roles})
	}
	err := db.Run(cmd, nil)
	if !isNoCmd(err) && user.UserSource != "" && (user.UserSource != "$external" || db.Name != "$external") {
		return fmt.Errorf("MongoDB 2.6+ does not support the UserSource setting")
	}
	return err
}

// AddUser creates or updates the authentication credentials of user within
// the db database.
//
// WARNING: This method is obsolete and should only be used with MongoDB 2.2
// or earlier. For MongoDB 2.4 and on, use UpsertUser instead.
func (db *Database) AddUser(username, password string, readOnly bool) error {
	// Try to emulate the old behavior on 2.6+
	user := &User{Username: username, Password: password}
	if db.Name == "admin" {
		if readOnly {
			user.Roles = []Role{RoleReadAny}
		} else {
			user.Roles = []Role{RoleReadWriteAny}
		}
	} else {
		if readOnly {
			user.Roles = []Role{RoleRead}
		} else {
			user.Roles = []Role{RoleReadWrite}
		}
	}
	err := db.runUserCmd("updateUser", user)
	if isNotFound(err) {
		return db.runUserCmd("createUser", user)
	}
	if !isNoCmd(err) {
		return err
	}

	// Command doesn't exist. Fallback to pre-2.6 behavior.
	psum := md5.New()
	psum.Write([]byte(username + ":mongo:" + password))
	digest := hex.EncodeToString(psum.Sum(nil))
	c := db.C("system.users")
	_, err = c.Upsert(bson.M{"user": username}, bson.M{"$set": bson.M{"user": username, "pwd": digest, "readOnly": readOnly}})
	return err
}

// RemoveUser removes the authentication credentials of user from the database.
func (db *Database) RemoveUser(user string) error {
	err := db.Run(bson.D{{"dropUser", user}}, nil)
	if isNoCmd(err) {
		users := db.C("system.users")
		return users.Remove(bson.M{"user": user})
	}
	if isNotFound(err) {
		return ErrNotFound
	}
	return err
}

type indexSpec struct {
	Name, NS         string
	Key              bson.D
	Unique           bool    ",omitempty"
	DropDups         bool    "dropDups,omitempty"
	Background       bool    ",omitempty"
	Sparse           bool    ",omitempty"
	Bits             int     ",omitempty"
	Min, Max         float64 ",omitempty"
	BucketSize       float64 "bucketSize,omitempty"
	ExpireAfter      int     "expireAfterSeconds,omitempty"
	Weights          bson.D  ",omitempty"
	DefaultLanguage  string  "default_language,omitempty"
	LanguageOverride string  "language_override,omitempty"
	TextIndexVersion int     "textIndexVersion,omitempty"
}

type Index struct {
	Key        []string // Index key fields; prefix name with dash (-) for descending order
	Unique     bool     // Prevent two documents from having the same index key
	DropDups   bool     // Drop documents with the same index key as a previously indexed one
	Background bool     // Build index in background and return immediately
	Sparse     bool     // Only index documents containing the Key fields

	// If ExpireAfter is defined the server will periodically delete
	// documents with indexed time.Time older than the provided delta.
	ExpireAfter time.Duration

	// Name holds the stored index name. On creation if this field is unset it is
	// computed by EnsureIndex based on the index key.
	Name string

	// Properties for spatial indexes.
	//
	// Min and Max were improperly typed as int when they should have been
	// floats.  To preserve backwards compatibility they are still typed as
	// int and the following two fields enable reading and writing the same
	// fields as float numbers. In mgo.v3, these fields will be dropped and
	// Min/Max will become floats.
	Min, Max   int
	Minf, Maxf float64
	BucketSize float64
	Bits       int

	// Properties for text indexes.
	DefaultLanguage  string
	LanguageOverride string

	// Weights defines the significance of provided fields relative to other
	// fields in a text index. The score for a given word in a document is derived
	// from the weighted sum of the frequency for each of the indexed fields in
	// that document. The default field weight is 1.
	Weights map[string]int
}

// mgo.v3: Drop Minf and Maxf and transform Min and Max to floats.
// mgo.v3: Drop DropDups as it's unsupported past 2.8.

type indexKeyInfo struct {
	name    string
	key     bson.D
	weights bson.D
}

func parseIndexKey(key []string) (*indexKeyInfo, error) {
	var keyInfo indexKeyInfo
	isText := false
	var order interface{}
	for _, field := range key {
		raw := field
		if keyInfo.name != "" {
			keyInfo.name += "_"
		}
		var kind string
		if field != "" {
			if field[0] == '$' {
				if c := strings.Index(field, ":"); c > 1 && c < len(field)-1 {
					kind = field[1:c]
					field = field[c+1:]
					keyInfo.name += field + "_" + kind
				} else {
					field = "\x00"
				}
			}
			switch field[0] {
			case 0:
				// Logic above failed. Reset and error.
				field = ""
			case '@':
				order = "2d"
				field = field[1:]
				// The shell used to render this field as key_ instead of key_2d,
				// and mgo followed suit. This has been fixed in recent server
				// releases, and mgo followed as well.
				keyInfo.name += field + "_2d"
			case '-':
				order = -1
				field = field[1:]
				keyInfo.name += field + "_-1"
			case '+':
				field = field[1:]
				fallthrough
			default:
				if kind == "" {
					order = 1
					keyInfo.name += field + "_1"
				} else {
					order = kind
				}
			}
		}
		if field == "" || kind != "" && order != kind {
			return nil, fmt.Errorf(`invalid index key: want "[$<kind>:][-]<field name>", got %q`, raw)
		}
		if kind == "text" {
			if !isText {
				keyInfo.key = append(keyInfo.key, bson.DocElem{"_fts", "text"}, bson.DocElem{"_ftsx", 1})
				isText = true
			}
			keyInfo.weights = append(keyInfo.weights, bson.DocElem{field, 1})
		} else {
			keyInfo.key = append(keyInfo.key, bson.DocElem{field, order})
		}
	}
	if keyInfo.name == "" {
		return nil, errors.New("invalid index key: no fields provided")
	}
	return &keyInfo, nil
}

// EnsureIndexKey ensures an index with the given key exists, creating it
// if necessary.
//
// This example:
//
//     err := collection.EnsureIndexKey("a", "b")
//
// Is equivalent to:
//
//     err := collection.EnsureIndex(mgo.Index{Key: []string{"a", "b"}})
//
// See the EnsureIndex method for more details.
func (c *Collection) EnsureIndexKey(key ...string) error {
	return c.EnsureIndex(Index{Key: key})
}

// EnsureIndex ensures an index with the given key exists, creating it with
// the provided parameters if necessary. EnsureIndex does not modify a previously
// existent index with a matching key. The old index must be dropped first instead.
//
// Once EnsureIndex returns successfully, following requests for the same index
// will not contact the server unless Collection.DropIndex is used to drop the
// same index, or Session.ResetIndexCache is called.
//
// For example:
//
//     index := Index{
//         Key: []string{"lastname", "firstname"},
//         Unique: true,
//         DropDups: true,
//         Background: true, // See notes.
//         Sparse: true,
//     }
//     err := collection.EnsureIndex(index)
//
// The Key value determines which fields compose the index. The index ordering
// will be ascending by default.  To obtain an index with a descending order,
// the field name should be prefixed by a dash (e.g. []string{"-time"}). It can
// also be optionally prefixed by an index kind, as in "$text:summary" or
// "$2d:-point". The key string format is:
//
//     [$<kind>:][-]<field name>
//
// If the Unique field is true, the index must necessarily contain only a single
// document per Key.  With DropDups set to true, documents with the same key
// as a previously indexed one will be dropped rather than an error returned.
//
// If Background is true, other connections will be allowed to proceed using
// the collection without the index while it's being built. Note that the
// session executing EnsureIndex will be blocked for as long as it takes for
// the index to be built.
//
// If Sparse is true, only documents containing the provided Key fields will be
// included in the index.  When using a sparse index for sorting, only indexed
// documents will be returned.
//
// If ExpireAfter is non-zero, the server will periodically scan the collection
// and remove documents containing an indexed time.Time field with a value
// older than ExpireAfter. See the documentation for details:
//
//     http://docs.mongodb.org/manual/tutorial/expire-data
//
// Other kinds of indexes are also supported through that API. Here is an example:
//
//     index := Index{
//         Key: []string{"$2d:loc"},
//         Bits: 26,
//     }
//     err := collection.EnsureIndex(index)
//
// The example above requests the creation of a "2d" index for the "loc" field.
//
// The 2D index bounds may be changed using the Min and Max attributes of the
// Index value.  The default bound setting of (-180, 180) is suitable for
// latitude/longitude pairs.
//
// The Bits parameter sets the precision of the 2D geohash values.  If not
// provided, 26 bits are used, which is roughly equivalent to 1 foot of
// precision for the default (-180, 180) index bounds.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Indexes
//     http://www.mongodb.org/display/DOCS/Indexing+Advice+and+FAQ
//     http://www.mongodb.org/display/DOCS/Indexing+as+a+Background+Operation
//     http://www.mongodb.org/display/DOCS/Geospatial+Indexing
//     http://www.mongodb.org/display/DOCS/Multikeys
//
func (c *Collection) EnsureIndex(index Index) error {
	keyInfo, err := parseIndexKey(index.Key)
	if err != nil {
		return err
	}

	session := c.Database.Session
	cacheKey := c.FullName + "\x00" + keyInfo.name
	if session.cluster().HasCachedIndex(cacheKey) {
		return nil
	}

	spec := indexSpec{
		Name:             keyInfo.name,
		NS:               c.FullName,
		Key:              keyInfo.key,
		Unique:           index.Unique,
		DropDups:         index.DropDups,
		Background:       index.Background,
		Sparse:           index.Sparse,
		Bits:             index.Bits,
		Min:              index.Minf,
		Max:              index.Maxf,
		BucketSize:       index.BucketSize,
		ExpireAfter:      int(index.ExpireAfter / time.Second),
		Weights:          keyInfo.weights,
		DefaultLanguage:  index.DefaultLanguage,
		LanguageOverride: index.LanguageOverride,
	}

	if spec.Min == 0 && spec.Max == 0 {
		spec.Min = float64(index.Min)
		spec.Max = float64(index.Max)
	}

	if index.Name != "" {
		spec.Name = index.Name
	}

NextField:
	for name, weight := range index.Weights {
		for i, elem := range spec.Weights {
			if elem.Name == name {
				spec.Weights[i].Value = weight
				continue NextField
			}
		}
		panic("weight provided for field that is not part of index key: " + name)
	}

	cloned := session.Clone()
	defer cloned.Close()
	cloned.SetMode(Strong, false)
	cloned.EnsureSafe(&Safe{})
	db := c.Database.With(cloned)

	// Try with a command first.
	err = db.Run(bson.D{{"createIndexes", c.Name}, {"indexes", []indexSpec{spec}}}, nil)
	if isNoCmd(err) {
		// Command not yet supported. Insert into the indexes collection instead.
		err = db.C("system.indexes").Insert(&spec)
	}
	if err == nil {
		session.cluster().CacheIndex(cacheKey, true)
	}
	return err
}

// DropIndex drops the index with the provided key from the c collection.
//
// See EnsureIndex for details on the accepted key variants.
//
// For example:
//
//     err1 := collection.DropIndex("firstField", "-secondField")
//     err2 := collection.DropIndex("customIndexName")
//
func (c *Collection) DropIndex(key ...string) error {
	keyInfo, err := parseIndexKey(key)
	if err != nil {
		return err
	}

	session := c.Database.Session
	cacheKey := c.FullName + "\x00" + keyInfo.name
	session.cluster().CacheIndex(cacheKey, false)

	session = session.Clone()
	defer session.Close()
	session.SetMode(Strong, false)

	db := c.Database.With(session)
	result := struct {
		ErrMsg string
		Ok     bool
	}{}
	err = db.Run(bson.D{{"dropIndexes", c.Name}, {"index", keyInfo.name}}, &result)
	if err != nil {
		return err
	}
	if !result.Ok {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// DropIndexName removes the index with the provided index name.
//
// For example:
//
//     err := collection.DropIndex("customIndexName")
//
func (c *Collection) DropIndexName(name string) error {
	session := c.Database.Session

	session = session.Clone()
	defer session.Close()
	session.SetMode(Strong, false)

	c = c.With(session)

	indexes, err := c.Indexes()
	if err != nil {
		return err
	}

	var index Index
	for _, idx := range indexes {
		if idx.Name == name {
			index = idx
			break
		}
	}

	if index.Name != "" {
		keyInfo, err := parseIndexKey(index.Key)
		if err != nil {
			return err
		}

		cacheKey := c.FullName + "\x00" + keyInfo.name
		session.cluster().CacheIndex(cacheKey, false)
	}

	result := struct {
		ErrMsg string
		Ok     bool
	}{}
	err = c.Database.Run(bson.D{{"dropIndexes", c.Name}, {"index", name}}, &result)
	if err != nil {
		return err
	}
	if !result.Ok {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// nonEventual returns a clone of session and ensures it is not Eventual.
// This guarantees that the server that is used for queries may be reused
// afterwards when a cursor is received.
func (session *Session) nonEventual() *Session {
	cloned := session.Clone()
	if cloned.consistency == Eventual {
		cloned.SetMode(Monotonic, false)
	}
	return cloned
}

// Indexes returns a list of all indexes for the collection.
//
// For example, this snippet would drop all available indexes:
//
//   indexes, err := collection.Indexes()
//   if err != nil {
//       return err
//   }
//   for _, index := range indexes {
//       err = collection.DropIndex(index.Key...)
//       if err != nil {
//           return err
//       }
//   }
//
// See the EnsureIndex method for more details on indexes.
func (c *Collection) Indexes() (indexes []Index, err error) {
	cloned := c.Database.Session.nonEventual()
	defer cloned.Close()

	batchSize := int(cloned.queryConfig.op.limit)

	// Try with a command.
	var result struct {
		Indexes []bson.Raw
		Cursor  cursorData
	}
	var iter *Iter
	err = c.Database.With(cloned).Run(bson.D{{"listIndexes", c.Name}, {"cursor", bson.D{{"batchSize", batchSize}}}}, &result)
	if err == nil {
		firstBatch := result.Indexes
		if firstBatch == nil {
			firstBatch = result.Cursor.FirstBatch
		}
		ns := strings.SplitN(result.Cursor.NS, ".", 2)
		if len(ns) < 2 {
			iter = c.With(cloned).NewIter(nil, firstBatch, result.Cursor.Id, nil)
		} else {
			iter = cloned.DB(ns[0]).C(ns[1]).NewIter(nil, firstBatch, result.Cursor.Id, nil)
		}
	} else if isNoCmd(err) {
		// Command not yet supported. Query the database instead.
		iter = c.Database.C("system.indexes").Find(bson.M{"ns": c.FullName}).Iter()
	} else {
		return nil, err
	}

	var spec indexSpec
	for iter.Next(&spec) {
		indexes = append(indexes, indexFromSpec(spec))
	}
	if err = iter.Close(); err != nil {
		return nil, err
	}
	sort.Sort(indexSlice(indexes))
	return indexes, nil
}

func indexFromSpec(spec indexSpec) Index {
	index := Index{
		Name:             spec.Name,
		Key:              simpleIndexKey(spec.Key),
		Unique:           spec.Unique,
		DropDups:         spec.DropDups,
		Background:       spec.Background,
		Sparse:           spec.Sparse,
		Minf:             spec.Min,
		Maxf:             spec.Max,
		Bits:             spec.Bits,
		BucketSize:       spec.BucketSize,
		DefaultLanguage:  spec.DefaultLanguage,
		LanguageOverride: spec.LanguageOverride,
		ExpireAfter:      time.Duration(spec.ExpireAfter) * time.Second,
	}
	if float64(int(spec.Min)) == spec.Min && float64(int(spec.Max)) == spec.Max {
		index.Min = int(spec.Min)
		index.Max = int(spec.Max)
	}
	if spec.TextIndexVersion > 0 {
		index.Key = make([]string, len(spec.Weights))
		index.Weights = make(map[string]int)
		for i, elem := range spec.Weights {
			index.Key[i] = "$text:" + elem.Name
			if w, ok := elem.Value.(int); ok {
				index.Weights[elem.Name] = w
			}
		}
	}
	return index
}

type indexSlice []Index

func (idxs indexSlice) Len() int           { return len(idxs) }
func (idxs indexSlice) Less(i, j int) bool { return idxs[i].Name < idxs[j].Name }
func (idxs indexSlice) Swap(i, j int)      { idxs[i], idxs[j] = idxs[j], idxs[i] }

func simpleIndexKey(realKey bson.D) (key []string) {
	for i := range realKey {
		field := realKey[i].Name
		vi, ok := realKey[i].Value.(int)
		if !ok {
			vf, _ := realKey[i].Value.(float64)
			vi = int(vf)
		}
		if vi == 1 {
			key = append(key, field)
			continue
		}
		if vi == -1 {
			key = append(key, "-"+field)
			continue
		}
		if vs, ok := realKey[i].Value.(string); ok {
			key = append(key, "$"+vs+":"+field)
			continue
		}
		panic("Got unknown index key type for field " + field)
	}
	return
}

// ResetIndexCache() clears the cache of previously ensured indexes.
// Following requests to EnsureIndex will contact the server.
func (s *Session) ResetIndexCache() {
	s.cluster().ResetIndexCache()
}

// New creates a new session with the same parameters as the original
// session, including consistency, batch size, prefetching, safety mode,
// etc. The returned session will use sockets from the pool, so there's
// a chance that writes just performed in another session may not yet
// be visible.
//
// Login information from the original session will not be copied over
// into the new session unless it was provided through the initial URL
// for the Dial function.
//
// See the Copy and Clone methods.
//
func (s *Session) New() *Session {
	s.m.Lock()
	scopy := copySession(s, false)
	s.m.Unlock()
	scopy.Refresh()
	return scopy
}

// Copy works just like New, but preserves the exact authentication
// information from the original session.
func (s *Session) Copy() *Session {
	s.m.Lock()
	scopy := copySession(s, true)
	s.m.Unlock()
	scopy.Refresh()
	return scopy
}

// Clone works just like Copy, but also reuses the same socket as the original
// session, in case it had already reserved one due to its consistency
// guarantees.  This behavior ensures that writes performed in the old session
// are necessarily observed when using the new session, as long as it was a
// strong or monotonic session.  That said, it also means that long operations
// may cause other goroutines using the original session to wait.
func (s *Session) Clone() *Session {
	s.m.Lock()
	scopy := copySession(s, true)
	s.m.Unlock()
	return scopy
}

// Close terminates the session.  It's a runtime error to use a session
// after it has been closed.
func (s *Session) Close() {
	s.m.Lock()
	if s.cluster_ != nil {
		debugf("Closing session %p", s)
		s.unsetSocket()
		s.cluster_.Release()
		s.cluster_ = nil
	}
	s.m.Unlock()
}

func (s *Session) cluster() *mongoCluster {
	if s.cluster_ == nil {
		panic("Session already closed")
	}
	return s.cluster_
}

// Refresh puts back any reserved sockets in use and restarts the consistency
// guarantees according to the current consistency setting for the session.
func (s *Session) Refresh() {
	s.m.Lock()
	s.slaveOk = s.consistency != Strong
	s.unsetSocket()
	s.m.Unlock()
}

// SetMode changes the consistency mode for the session.
//
// In the Strong consistency mode reads and writes will always be made to
// the primary server using a unique connection so that reads and writes are
// fully consistent, ordered, and observing the most up-to-date data.
// This offers the least benefits in terms of distributing load, but the
// most guarantees.  See also Monotonic and Eventual.
//
// In the Monotonic consistency mode reads may not be entirely up-to-date,
// but they will always see the history of changes moving forward, the data
// read will be consistent across sequential queries in the same session,
// and modifications made within the session will be observed in following
// queries (read-your-writes).
//
// In practice, the Monotonic mode is obtained by performing initial reads
// on a unique connection to an arbitrary secondary, if one is available,
// and once the first write happens, the session connection is switched over
// to the primary server.  This manages to distribute some of the reading
// load with secondaries, while maintaining some useful guarantees.
//
// In the Eventual consistency mode reads will be made to any secondary in the
// cluster, if one is available, and sequential reads will not necessarily
// be made with the same connection.  This means that data may be observed
// out of order.  Writes will of course be issued to the primary, but
// independent writes in the same Eventual session may also be made with
// independent connections, so there are also no guarantees in terms of
// write ordering (no read-your-writes guarantees either).
//
// The Eventual mode is the fastest and most resource-friendly, but is
// also the one offering the least guarantees about ordering of the data
// read and written.
//
// If refresh is true, in addition to ensuring the session is in the given
// consistency mode, the consistency guarantees will also be reset (e.g.
// a Monotonic session will be allowed to read from secondaries again).
// This is equivalent to calling the Refresh function.
//
// Shifting between Monotonic and Strong modes will keep a previously
// reserved connection for the session unless refresh is true or the
// connection is unsuitable (to a secondary server in a Strong session).
func (s *Session) SetMode(consistency Mode, refresh bool) {
	s.m.Lock()
	debugf("Session %p: setting mode %d with refresh=%v (master=%p, slave=%p)", s, consistency, refresh, s.masterSocket, s.slaveSocket)
	s.consistency = consistency
	if refresh {
		s.slaveOk = s.consistency != Strong
		s.unsetSocket()
	} else if s.consistency == Strong {
		s.slaveOk = false
	} else if s.masterSocket == nil {
		s.slaveOk = true
	}
	s.m.Unlock()
}

// Mode returns the current consistency mode for the session.
func (s *Session) Mode() Mode {
	s.m.RLock()
	mode := s.consistency
	s.m.RUnlock()
	return mode
}

// SetSyncTimeout sets the amount of time an operation with this session
// will wait before returning an error in case a connection to a usable
// server can't be established. Set it to zero to wait forever. The
// default value is 7 seconds.
func (s *Session) SetSyncTimeout(d time.Duration) {
	s.m.Lock()
	s.syncTimeout = d
	s.m.Unlock()
}

// SetSocketTimeout sets the amount of time to wait for a non-responding
// socket to the database before it is forcefully closed.
func (s *Session) SetSocketTimeout(d time.Duration) {
	s.m.Lock()
	s.sockTimeout = d
	if s.masterSocket != nil {
		s.masterSocket.SetTimeout(d)
	}
	if s.slaveSocket != nil {
		s.slaveSocket.SetTimeout(d)
	}
	s.m.Unlock()
}

// SetCursorTimeout changes the standard timeout period that the server
// enforces on created cursors. The only supported value right now is
// 0, which disables the timeout. The standard server timeout is 10 minutes.
func (s *Session) SetCursorTimeout(d time.Duration) {
	s.m.Lock()
	if d == 0 {
		s.queryConfig.op.flags |= flagNoCursorTimeout
	} else {
		panic("SetCursorTimeout: only 0 (disable timeout) supported for now")
	}
	s.m.Unlock()
}

// SetPoolLimit sets the maximum number of sockets in use in a single server
// before this session will block waiting for a socket to be available.
// The default limit is 4096.
//
// This limit must be set to cover more than any expected workload of the
// application. It is a bad practice and an unsupported use case to use the
// database driver to define the concurrency limit of an application. Prevent
// such concurrency "at the door" instead, by properly restricting the amount
// of used resources and number of goroutines before they are created.
func (s *Session) SetPoolLimit(limit int) {
	s.m.Lock()
	s.poolLimit = limit
	s.m.Unlock()
}

// SetBypassValidation sets whether the server should bypass the registered
// validation expressions executed when documents are inserted or modified,
// in the interest of preserving invariants in the collection being modified.
// The default is to not bypass, and thus to perform the validation
// expressions registered for modified collections.
//
// Document validation was introuced in MongoDB 3.2.
//
// Relevant documentation:
//
//   https://docs.mongodb.org/manual/release-notes/3.2/#bypass-validation
//
func (s *Session) SetBypassValidation(bypass bool) {
	s.m.Lock()
	s.bypassValidation = bypass
	s.m.Unlock()
}

// SetBatch sets the default batch size used when fetching documents from the
// database. It's possible to change this setting on a per-query basis as
// well, using the Query.Batch method.
//
// The default batch size is defined by the database itself.  As of this
// writing, MongoDB will use an initial size of min(100 docs, 4MB) on the
// first batch, and 4MB on remaining ones.
func (s *Session) SetBatch(n int) {
	if n == 1 {
		// Server interprets 1 as -1 and closes the cursor (!?)
		n = 2
	}
	s.m.Lock()
	s.queryConfig.op.limit = int32(n)
	s.m.Unlock()
}

// SetPrefetch sets the default point at which the next batch of results will be
// requested.  When there are p*batch_size remaining documents cached in an
// Iter, the next batch will be requested in background. For instance, when
// using this:
//
//     session.SetBatch(200)
//     session.SetPrefetch(0.25)
//
// and there are only 50 documents cached in the Iter to be processed, the
// next batch of 200 will be requested. It's possible to change this setting on
// a per-query basis as well, using the Prefetch method of Query.
//
// The default prefetch value is 0.25.
func (s *Session) SetPrefetch(p float64) {
	s.m.Lock()
	s.queryConfig.prefetch = p
	s.m.Unlock()
}

// See SetSafe for details on the Safe type.
type Safe struct {
	W        int    // Min # of servers to ack before success
	WMode    string // Write mode for MongoDB 2.0+ (e.g. "majority")
	WTimeout int    // Milliseconds to wait for W before timing out
	FSync    bool   // Sync via the journal if present, or via data files sync otherwise
	J        bool   // Sync via the journal if present
}

// Safe returns the current safety mode for the session.
func (s *Session) Safe() (safe *Safe) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.safeOp != nil {
		cmd := s.safeOp.query.(*getLastError)
		safe = &Safe{WTimeout: cmd.WTimeout, FSync: cmd.FSync, J: cmd.J}
		switch w := cmd.W.(type) {
		case string:
			safe.WMode = w
		case int:
			safe.W = w
		}
	}
	return
}

// SetSafe changes the session safety mode.
//
// If the safe parameter is nil, the session is put in unsafe mode, and writes
// become fire-and-forget, without error checking.  The unsafe mode is faster
// since operations won't hold on waiting for a confirmation.
//
// If the safe parameter is not nil, any changing query (insert, update, ...)
// will be followed by a getLastError command with the specified parameters,
// to ensure the request was correctly processed.
//
// The safe.W parameter determines how many servers should confirm a write
// before the operation is considered successful.  If set to 0 or 1, the
// command will return as soon as the primary is done with the request.
// If safe.WTimeout is greater than zero, it determines how many milliseconds
// to wait for the safe.W servers to respond before returning an error.
//
// Starting with MongoDB 2.0.0 the safe.WMode parameter can be used instead
// of W to request for richer semantics. If set to "majority" the server will
// wait for a majority of members from the replica set to respond before
// returning. Custom modes may also be defined within the server to create
// very detailed placement schemas. See the data awareness documentation in
// the links below for more details (note that MongoDB internally reuses the
// "w" field name for WMode).
//
// If safe.J is true, servers will block until write operations have been
// committed to the journal. Cannot be used in combination with FSync. Prior
// to MongoDB 2.6 this option was ignored if the server was running without
// journaling. Starting with MongoDB 2.6 write operations will fail with an
// exception if this option is used when the server is running without
// journaling.
//
// If safe.FSync is true and the server is running without journaling, blocks
// until the server has synced all data files to disk. If the server is running
// with journaling, this acts the same as the J option, blocking until write
// operations have been committed to the journal. Cannot be used in
// combination with J.
//
// Since MongoDB 2.0.0, the safe.J option can also be used instead of FSync
// to force the server to wait for a group commit in case journaling is
// enabled. The option has no effect if the server has journaling disabled.
//
// For example, the following statement will make the session check for
// errors, without imposing further constraints:
//
//     session.SetSafe(&mgo.Safe{})
//
// The following statement will force the server to wait for a majority of
// members of a replica set to return (MongoDB 2.0+ only):
//
//     session.SetSafe(&mgo.Safe{WMode: "majority"})
//
// The following statement, on the other hand, ensures that at least two
// servers have flushed the change to disk before confirming the success
// of operations:
//
//     session.EnsureSafe(&mgo.Safe{W: 2, FSync: true})
//
// The following statement, on the other hand, disables the verification
// of errors entirely:
//
//     session.SetSafe(nil)
//
// See also the EnsureSafe method.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/getLastError+Command
//     http://www.mongodb.org/display/DOCS/Verifying+Propagation+of+Writes+with+getLastError
//     http://www.mongodb.org/display/DOCS/Data+Center+Awareness
//
func (s *Session) SetSafe(safe *Safe) {
	s.m.Lock()
	s.safeOp = nil
	s.ensureSafe(safe)
	s.m.Unlock()
}

// EnsureSafe compares the provided safety parameters with the ones
// currently in use by the session and picks the most conservative
// choice for each setting.
//
// That is:
//
//     - safe.WMode is always used if set.
//     - safe.W is used if larger than the current W and WMode is empty.
//     - safe.FSync is always used if true.
//     - safe.J is used if FSync is false.
//     - safe.WTimeout is used if set and smaller than the current WTimeout.
//
// For example, the following statement will ensure the session is
// at least checking for errors, without enforcing further constraints.
// If a more conservative SetSafe or EnsureSafe call was previously done,
// the following call will be ignored.
//
//     session.EnsureSafe(&mgo.Safe{})
//
// See also the SetSafe method for details on what each option means.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/getLastError+Command
//     http://www.mongodb.org/display/DOCS/Verifying+Propagation+of+Writes+with+getLastError
//     http://www.mongodb.org/display/DOCS/Data+Center+Awareness
//
func (s *Session) EnsureSafe(safe *Safe) {
	s.m.Lock()
	s.ensureSafe(safe)
	s.m.Unlock()
}

func (s *Session) ensureSafe(safe *Safe) {
	if safe == nil {
		return
	}

	var w interface{}
	if safe.WMode != "" {
		w = safe.WMode
	} else if safe.W > 0 {
		w = safe.W
	}

	var cmd getLastError
	if s.safeOp == nil {
		cmd = getLastError{1, w, safe.WTimeout, safe.FSync, safe.J}
	} else {
		// Copy.  We don't want to mutate the existing query.
		cmd = *(s.safeOp.query.(*getLastError))
		if cmd.W == nil {
			cmd.W = w
		} else if safe.WMode != "" {
			cmd.W = safe.WMode
		} else if i, ok := cmd.W.(int); ok && safe.W > i {
			cmd.W = safe.W
		}
		if safe.WTimeout > 0 && safe.WTimeout < cmd.WTimeout {
			cmd.WTimeout = safe.WTimeout
		}
		if safe.FSync {
			cmd.FSync = true
			cmd.J = false
		} else if safe.J && !cmd.FSync {
			cmd.J = true
		}
	}
	s.safeOp = &queryOp{
		query:      &cmd,
		collection: "admin.$cmd",
		limit:      -1,
	}
}

// Run issues the provided command on the "admin" database and
// and unmarshals its result in the respective argument. The cmd
// argument may be either a string with the command name itself, in
// which case an empty document of the form bson.M{cmd: 1} will be used,
// or it may be a full command document.
//
// Note that MongoDB considers the first marshalled key as the command
// name, so when providing a command with options, it's important to
// use an ordering-preserving document, such as a struct value or an
// instance of bson.D.  For instance:
//
//     db.Run(bson.D{{"create", "mycollection"}, {"size", 1024}})
//
// For commands on arbitrary databases, see the Run method in
// the Database type.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Commands
//     http://www.mongodb.org/display/DOCS/List+of+Database+CommandSkips
//
func (s *Session) Run(cmd interface{}, result interface{}) error {
	return s.DB("admin").Run(cmd, result)
}

// SelectServers restricts communication to servers configured with the
// given tags. For example, the following statement restricts servers
// used for reading operations to those with both tag "disk" set to
// "ssd" and tag "rack" set to 1:
//
//     session.SelectServers(bson.D{{"disk", "ssd"}, {"rack", 1}})
//
// Multiple sets of tags may be provided, in which case the used server
// must match all tags within any one set.
//
// If a connection was previously assigned to the session due to the
// current session mode (see Session.SetMode), the tag selection will
// only be enforced after the session is refreshed.
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/tutorial/configure-replica-set-tag-sets
//
func (s *Session) SelectServers(tags ...bson.D) {
	s.m.Lock()
	s.queryConfig.op.serverTags = tags
	s.m.Unlock()
}

// Ping runs a trivial ping command just to get in touch with the server.
func (s *Session) Ping() error {
	return s.Run("ping", nil)
}

// Fsync flushes in-memory writes to disk on the server the session
// is established with. If async is true, the call returns immediately,
// otherwise it returns after the flush has been made.
func (s *Session) Fsync(async bool) error {
	return s.Run(bson.D{{"fsync", 1}, {"async", async}}, nil)
}

// FsyncLock locks all writes in the specific server the session is
// established with and returns. Any writes attempted to the server
// after it is successfully locked will block until FsyncUnlock is
// called for the same server.
//
// This method works on secondaries as well, preventing the oplog from
// being flushed while the server is locked, but since only the server
// connected to is locked, for locking specific secondaries it may be
// necessary to establish a connection directly to the secondary (see
// Dial's connect=direct option).
//
// As an important caveat, note that once a write is attempted and
// blocks, follow up reads will block as well due to the way the
// lock is internally implemented in the server. More details at:
//
//     https://jira.mongodb.org/browse/SERVER-4243
//
// FsyncLock is often used for performing consistent backups of
// the database files on disk.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/fsync+Command
//     http://www.mongodb.org/display/DOCS/Backups
//
func (s *Session) FsyncLock() error {
	return s.Run(bson.D{{"fsync", 1}, {"lock", true}}, nil)
}

// FsyncUnlock releases the server for writes. See FsyncLock for details.
func (s *Session) FsyncUnlock() error {
	err := s.Run(bson.D{{"fsyncUnlock", 1}}, nil)
	if isNoCmd(err) {
		err = s.DB("admin").C("$cmd.sys.unlock").Find(nil).One(nil) // WTF?
	}
	return err
}

// Find prepares a query using the provided document.  The document may be a
// map or a struct value capable of being marshalled with bson.  The map
// may be a generic one using interface{} for its key and/or values, such as
// bson.M, or it may be a properly typed map.  Providing nil as the document
// is equivalent to providing an empty document such as bson.M{}.
//
// Further details of the query may be tweaked using the resulting Query value,
// and then executed to retrieve results using methods such as One, For,
// Iter, or Tail.
//
// In case the resulting document includes a field named $err or errmsg, which
// are standard ways for MongoDB to return query errors, the returned err will
// be set to a *QueryError value including the Err message and the Code.  In
// those cases, the result argument is still unmarshalled into with the
// received document so that any other custom values may be obtained if
// desired.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Querying
//     http://www.mongodb.org/display/DOCS/Advanced+Queries
//
func (c *Collection) Find(query interface{}) *Query {
	session := c.Database.Session
	session.m.RLock()
	q := &Query{session: session, query: session.queryConfig}
	session.m.RUnlock()
	q.op.query = query
	q.op.collection = c.FullName
	return q
}

type repairCmd struct {
	RepairCursor string           `bson:"repairCursor"`
	Cursor       *repairCmdCursor ",omitempty"
}

type repairCmdCursor struct {
	BatchSize int `bson:"batchSize,omitempty"`
}

// Repair returns an iterator that goes over all recovered documents in the
// collection, in a best-effort manner. This is most useful when there are
// damaged data files. Multiple copies of the same document may be returned
// by the iterator.
//
// Repair is supported in MongoDB 2.7.8 and later.
func (c *Collection) Repair() *Iter {
	// Clone session and set it to Monotonic mode so that the server
	// used for the query may be safely obtained afterwards, if
	// necessary for iteration when a cursor is received.
	session := c.Database.Session
	cloned := session.nonEventual()
	defer cloned.Close()

	batchSize := int(cloned.queryConfig.op.limit)

	var result struct{ Cursor cursorData }

	cmd := repairCmd{
		RepairCursor: c.Name,
		Cursor:       &repairCmdCursor{batchSize},
	}

	clonedc := c.With(cloned)
	err := clonedc.Database.Run(cmd, &result)
	return clonedc.NewIter(session, result.Cursor.FirstBatch, result.Cursor.Id, err)
}

// FindId is a convenience helper equivalent to:
//
//     query := collection.Find(bson.M{"_id": id})
//
// See the Find method for more details.
func (c *Collection) FindId(id interface{}) *Query {
	return c.Find(bson.D{{"_id", id}})
}

type Pipe struct {
	session    *Session
	collection *Collection
	pipeline   interface{}
	allowDisk  bool
	batchSize  int
}

type pipeCmd struct {
	Aggregate string
	Pipeline  interface{}
	Cursor    *pipeCmdCursor ",omitempty"
	Explain   bool           ",omitempty"
	AllowDisk bool           "allowDiskUse,omitempty"
}

type pipeCmdCursor struct {
	BatchSize int `bson:"batchSize,omitempty"`
}

// Pipe prepares a pipeline to aggregate. The pipeline document
// must be a slice built in terms of the aggregation framework language.
//
// For example:
//
//     pipe := collection.Pipe([]bson.M{{"$match": bson.M{"name": "Otavio"}}})
//     iter := pipe.Iter()
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/reference/aggregation
//     http://docs.mongodb.org/manual/applications/aggregation
//     http://docs.mongodb.org/manual/tutorial/aggregation-examples
//
func (c *Collection) Pipe(pipeline interface{}) *Pipe {
	session := c.Database.Session
	session.m.RLock()
	batchSize := int(session.queryConfig.op.limit)
	session.m.RUnlock()
	return &Pipe{
		session:    session,
		collection: c,
		pipeline:   pipeline,
		batchSize:  batchSize,
	}
}

// Iter executes the pipeline and returns an iterator capable of going
// over all the generated results.
func (p *Pipe) Iter() *Iter {
	// Clone session and set it to Monotonic mode so that the server
	// used for the query may be safely obtained afterwards, if
	// necessary for iteration when a cursor is received.
	cloned := p.session.nonEventual()
	defer cloned.Close()
	c := p.collection.With(cloned)

	var result struct {
		Result []bson.Raw // 2.4, no cursors.
		Cursor cursorData // 2.6+, with cursors.
	}

	cmd := pipeCmd{
		Aggregate: c.Name,
		Pipeline:  p.pipeline,
		AllowDisk: p.allowDisk,
		Cursor:    &pipeCmdCursor{p.batchSize},
	}
	err := c.Database.Run(cmd, &result)
	if e, ok := err.(*QueryError); ok && e.Message == `unrecognized field "cursor` {
		cmd.Cursor = nil
		cmd.AllowDisk = false
		err = c.Database.Run(cmd, &result)
	}
	firstBatch := result.Result
	if firstBatch == nil {
		firstBatch = result.Cursor.FirstBatch
	}
	return c.NewIter(p.session, firstBatch, result.Cursor.Id, err)
}

// NewIter returns a newly created iterator with the provided parameters.
// Using this method is not recommended unless the desired functionality
// is not yet exposed via a more convenient interface (Find, Pipe, etc).
//
// The optional session parameter associates the lifetime of the returned
// iterator to an arbitrary session. If nil, the iterator will be bound to
// c's session.
//
// Documents in firstBatch will be individually provided by the returned
// iterator before documents from cursorId are made available. If cursorId
// is zero, only the documents in firstBatch are provided.
//
// If err is not nil, the iterator's Err method will report it after
// exhausting documents in firstBatch.
//
// NewIter must be called right after the cursor id is obtained, and must not
// be called on a collection in Eventual mode, because the cursor id is
// associated with the specific server that returned it. The provided session
// parameter may be in any mode or state, though.
//
func (c *Collection) NewIter(session *Session, firstBatch []bson.Raw, cursorId int64, err error) *Iter {
	var server *mongoServer
	csession := c.Database.Session
	csession.m.RLock()
	socket := csession.masterSocket
	if socket == nil {
		socket = csession.slaveSocket
	}
	if socket != nil {
		server = socket.Server()
	}
	csession.m.RUnlock()

	if server == nil {
		if csession.Mode() == Eventual {
			panic("Collection.NewIter called in Eventual mode")
		}
		if err == nil {
			err = errors.New("server not available")
		}
	}

	if session == nil {
		session = csession
	}

	iter := &Iter{
		session: session,
		server:  server,
		timeout: -1,
		err:     err,
	}
	iter.gotReply.L = &iter.m
	for _, doc := range firstBatch {
		iter.docData.Push(doc.Data)
	}
	if cursorId != 0 {
		iter.op.cursorId = cursorId
		iter.op.collection = c.FullName
		iter.op.replyFunc = iter.replyFunc()
	}
	return iter
}

// All works like Iter.All.
func (p *Pipe) All(result interface{}) error {
	return p.Iter().All(result)
}

// One executes the pipeline and unmarshals the first item from the
// result set into the result parameter.
// It returns ErrNotFound if no items are generated by the pipeline.
func (p *Pipe) One(result interface{}) error {
	iter := p.Iter()
	if iter.Next(result) {
		return nil
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return ErrNotFound
}

// Explain returns a number of details about how the MongoDB server would
// execute the requested pipeline, such as the number of objects examined,
// the number of times the read lock was yielded to allow writes to go in,
// and so on.
//
// For example:
//
//     var m bson.M
//     err := collection.Pipe(pipeline).Explain(&m)
//     if err == nil {
//         fmt.Printf("Explain: %#v\n", m)
//     }
//
func (p *Pipe) Explain(result interface{}) error {
	c := p.collection
	cmd := pipeCmd{
		Aggregate: c.Name,
		Pipeline:  p.pipeline,
		AllowDisk: p.allowDisk,
		Explain:   true,
	}
	return c.Database.Run(cmd, result)
}

// AllowDiskUse enables writing to the "<dbpath>/_tmp" server directory so
// that aggregation pipelines do not have to be held entirely in memory.
func (p *Pipe) AllowDiskUse() *Pipe {
	p.allowDisk = true
	return p
}

// Batch sets the batch size used when fetching documents from the database.
// It's possible to change this setting on a per-session basis as well, using
// the Batch method of Session.
//
// The default batch size is defined by the database server.
func (p *Pipe) Batch(n int) *Pipe {
	p.batchSize = n
	return p
}

// mgo.v3: Use a single user-visible error type.

type LastError struct {
	Err             string
	Code, N, Waited int
	FSyncFiles      int `bson:"fsyncFiles"`
	WTimeout        bool
	UpdatedExisting bool        `bson:"updatedExisting"`
	UpsertedId      interface{} `bson:"upserted"`

	modified int
	ecases   []BulkErrorCase
}

func (err *LastError) Error() string {
	return err.Err
}

type queryError struct {
	Err           string "$err"
	ErrMsg        string
	Assertion     string
	Code          int
	AssertionCode int        "assertionCode"
	LastError     *LastError "lastErrorObject"
}

type QueryError struct {
	Code      int
	Message   string
	Assertion bool
}

func (err *QueryError) Error() string {
	return err.Message
}

// IsDup returns whether err informs of a duplicate key error because
// a primary key index or a secondary unique index already has an entry
// with the given value.
func IsDup(err error) bool {
	// Besides being handy, helps with MongoDB bugs SERVER-7164 and SERVER-11493.
	// What follows makes me sad. Hopefully conventions will be more clear over time.
	switch e := err.(type) {
	case *LastError:
		return e.Code == 11000 || e.Code == 11001 || e.Code == 12582 || e.Code == 16460 && strings.Contains(e.Err, " E11000 ")
	case *QueryError:
		return e.Code == 11000 || e.Code == 11001 || e.Code == 12582
	case *BulkError:
		for _, ecase := range e.ecases {
			if !IsDup(ecase.Err) {
				return false
			}
		}
		return true
	}
	return false
}

// Insert inserts one or more documents in the respective collection.  In
// case the session is in safe mode (see the SetSafe method) and an error
// happens while inserting the provided documents, the returned error will
// be of type *LastError.
func (c *Collection) Insert(docs ...interface{}) error {
	_, err := c.writeOp(&insertOp{c.FullName, docs, 0}, true)
	return err
}

// Update finds a single document matching the provided selector document
// and modifies it according to the update document.
// If the session is in safe mode (see SetSafe) a ErrNotFound error is
// returned if a document isn't found, or a value of type *LastError
// when some other error is detected.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Updating
//     http://www.mongodb.org/display/DOCS/Atomic+Operations
//
func (c *Collection) Update(selector interface{}, update interface{}) error {
	if selector == nil {
		selector = bson.D{}
	}
	op := updateOp{
		Collection: c.FullName,
		Selector:   selector,
		Update:     update,
	}
	lerr, err := c.writeOp(&op, true)
	if err == nil && lerr != nil && !lerr.UpdatedExisting {
		return ErrNotFound
	}
	return err
}

// UpdateId is a convenience helper equivalent to:
//
//     err := collection.Update(bson.M{"_id": id}, update)
//
// See the Update method for more details.
func (c *Collection) UpdateId(id interface{}, update interface{}) error {
	return c.Update(bson.D{{"_id", id}}, update)
}

// ChangeInfo holds details about the outcome of an update operation.
type ChangeInfo struct {
	// Updated reports the number of existing documents modified.
	// Due to server limitations, this reports the same value as the Matched field when
	// talking to MongoDB <= 2.4 and on Upsert and Apply (findAndModify) operations.
	Updated    int
	Removed    int         // Number of documents removed
	Matched    int         // Number of documents matched but not necessarily changed
	UpsertedId interface{} // Upserted _id field, when not explicitly provided
}

// UpdateAll finds all documents matching the provided selector document
// and modifies them according to the update document.
// If the session is in safe mode (see SetSafe) details of the executed
// operation are returned in info or an error of type *LastError when
// some problem is detected. It is not an error for the update to not be
// applied on any documents because the selector doesn't match.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Updating
//     http://www.mongodb.org/display/DOCS/Atomic+Operations
//
func (c *Collection) UpdateAll(selector interface{}, update interface{}) (info *ChangeInfo, err error) {
	if selector == nil {
		selector = bson.D{}
	}
	op := updateOp{
		Collection: c.FullName,
		Selector:   selector,
		Update:     update,
		Flags:      2,
		Multi:      true,
	}
	lerr, err := c.writeOp(&op, true)
	if err == nil && lerr != nil {
		info = &ChangeInfo{Updated: lerr.modified, Matched: lerr.N}
	}
	return info, err
}

// Upsert finds a single document matching the provided selector document
// and modifies it according to the update document.  If no document matching
// the selector is found, the update document is applied to the selector
// document and the result is inserted in the collection.
// If the session is in safe mode (see SetSafe) details of the executed
// operation are returned in info, or an error of type *LastError when
// some problem is detected.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Updating
//     http://www.mongodb.org/display/DOCS/Atomic+Operations
//
func (c *Collection) Upsert(selector interface{}, update interface{}) (info *ChangeInfo, err error) {
	if selector == nil {
		selector = bson.D{}
	}
	op := updateOp{
		Collection: c.FullName,
		Selector:   selector,
		Update:     update,
		Flags:      1,
		Upsert:     true,
	}
	lerr, err := c.writeOp(&op, true)
	if err == nil && lerr != nil {
		info = &ChangeInfo{}
		if lerr.UpdatedExisting {
			info.Matched = lerr.N
			info.Updated = lerr.modified
		} else {
			info.UpsertedId = lerr.UpsertedId
		}
	}
	return info, err
}

// UpsertId is a convenience helper equivalent to:
//
//     info, err := collection.Upsert(bson.M{"_id": id}, update)
//
// See the Upsert method for more details.
func (c *Collection) UpsertId(id interface{}, update interface{}) (info *ChangeInfo, err error) {
	return c.Upsert(bson.D{{"_id", id}}, update)
}

// Remove finds a single document matching the provided selector document
// and removes it from the database.
// If the session is in safe mode (see SetSafe) a ErrNotFound error is
// returned if a document isn't found, or a value of type *LastError
// when some other error is detected.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Removing
//
func (c *Collection) Remove(selector interface{}) error {
	if selector == nil {
		selector = bson.D{}
	}
	lerr, err := c.writeOp(&deleteOp{c.FullName, selector, 1, 1}, true)
	if err == nil && lerr != nil && lerr.N == 0 {
		return ErrNotFound
	}
	return err
}

// RemoveId is a convenience helper equivalent to:
//
//     err := collection.Remove(bson.M{"_id": id})
//
// See the Remove method for more details.
func (c *Collection) RemoveId(id interface{}) error {
	return c.Remove(bson.D{{"_id", id}})
}

// RemoveAll finds all documents matching the provided selector document
// and removes them from the database.  In case the session is in safe mode
// (see the SetSafe method) and an error happens when attempting the change,
// the returned error will be of type *LastError.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Removing
//
func (c *Collection) RemoveAll(selector interface{}) (info *ChangeInfo, err error) {
	if selector == nil {
		selector = bson.D{}
	}
	lerr, err := c.writeOp(&deleteOp{c.FullName, selector, 0, 0}, true)
	if err == nil && lerr != nil {
		info = &ChangeInfo{Removed: lerr.N, Matched: lerr.N}
	}
	return info, err
}

// DropDatabase removes the entire database including all of its collections.
func (db *Database) DropDatabase() error {
	return db.Run(bson.D{{"dropDatabase", 1}}, nil)
}

// DropCollection removes the entire collection including all of its documents.
func (c *Collection) DropCollection() error {
	return c.Database.Run(bson.D{{"drop", c.Name}}, nil)
}

// The CollectionInfo type holds metadata about a collection.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/createCollection+Command
//     http://www.mongodb.org/display/DOCS/Capped+Collections
//
type CollectionInfo struct {
	// DisableIdIndex prevents the automatic creation of the index
	// on the _id field for the collection.
	DisableIdIndex bool

	// ForceIdIndex enforces the automatic creation of the index
	// on the _id field for the collection. Capped collections,
	// for example, do not have such an index by default.
	ForceIdIndex bool

	// If Capped is true new documents will replace old ones when
	// the collection is full. MaxBytes must necessarily be set
	// to define the size when the collection wraps around.
	// MaxDocs optionally defines the number of documents when it
	// wraps, but MaxBytes still needs to be set.
	Capped   bool
	MaxBytes int
	MaxDocs  int

	// Validator contains a validation expression that defines which
	// documents should be considered valid for this collection.
	Validator interface{}

	// ValidationLevel may be set to "strict" (the default) to force
	// MongoDB to validate all documents on inserts and updates, to
	// "moderate" to apply the validation rules only to documents
	// that already fulfill the validation criteria, or to "off" for
	// disabling validation entirely.
	ValidationLevel string

	// ValidationAction determines how MongoDB handles documents that
	// violate the validation rules. It may be set to "error" (the default)
	// to reject inserts or updates that violate the rules, or to "warn"
	// to log invalid operations but allow them to proceed.
	ValidationAction string

	// StorageEngine allows specifying collection options for the
	// storage engine in use. The map keys must hold the storage engine
	// name for which options are being specified.
	StorageEngine interface{}
}

// Create explicitly creates the c collection with details of info.
// MongoDB creates collections automatically on use, so this method
// is only necessary when creating collection with non-default
// characteristics, such as capped collections.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/createCollection+Command
//     http://www.mongodb.org/display/DOCS/Capped+Collections
//
func (c *Collection) Create(info *CollectionInfo) error {
	cmd := make(bson.D, 0, 4)
	cmd = append(cmd, bson.DocElem{"create", c.Name})
	if info.Capped {
		if info.MaxBytes < 1 {
			return fmt.Errorf("Collection.Create: with Capped, MaxBytes must also be set")
		}
		cmd = append(cmd, bson.DocElem{"capped", true})
		cmd = append(cmd, bson.DocElem{"size", info.MaxBytes})
		if info.MaxDocs > 0 {
			cmd = append(cmd, bson.DocElem{"max", info.MaxDocs})
		}
	}
	if info.DisableIdIndex {
		cmd = append(cmd, bson.DocElem{"autoIndexId", false})
	}
	if info.ForceIdIndex {
		cmd = append(cmd, bson.DocElem{"autoIndexId", true})
	}
	if info.Validator != nil {
		cmd = append(cmd, bson.DocElem{"validator", info.Validator})
	}
	if info.ValidationLevel != "" {
		cmd = append(cmd, bson.DocElem{"validationLevel", info.ValidationLevel})
	}
	if info.ValidationAction != "" {
		cmd = append(cmd, bson.DocElem{"validationAction", info.ValidationAction})
	}
	if info.StorageEngine != nil {
		cmd = append(cmd, bson.DocElem{"storageEngine", info.StorageEngine})
	}
	return c.Database.Run(cmd, nil)
}

// Batch sets the batch size used when fetching documents from the database.
// It's possible to change this setting on a per-session basis as well, using
// the Batch method of Session.

// The default batch size is defined by the database itself.  As of this
// writing, MongoDB will use an initial size of min(100 docs, 4MB) on the
// first batch, and 4MB on remaining ones.
func (q *Query) Batch(n int) *Query {
	if n == 1 {
		// Server interprets 1 as -1 and closes the cursor (!?)
		n = 2
	}
	q.m.Lock()
	q.op.limit = int32(n)
	q.m.Unlock()
	return q
}

// Prefetch sets the point at which the next batch of results will be requested.
// When there are p*batch_size remaining documents cached in an Iter, the next
// batch will be requested in background. For instance, when using this:
//
//     query.Batch(200).Prefetch(0.25)
//
// and there are only 50 documents cached in the Iter to be processed, the
// next batch of 200 will be requested. It's possible to change this setting on
// a per-session basis as well, using the SetPrefetch method of Session.
//
// The default prefetch value is 0.25.
func (q *Query) Prefetch(p float64) *Query {
	q.m.Lock()
	q.prefetch = p
	q.m.Unlock()
	return q
}

// Skip skips over the n initial documents from the query results.  Note that
// this only makes sense with capped collections where documents are naturally
// ordered by insertion time, or with sorted results.
func (q *Query) Skip(n int) *Query {
	q.m.Lock()
	q.op.skip = int32(n)
	q.m.Unlock()
	return q
}

// Limit restricts the maximum number of documents retrieved to n, and also
// changes the batch size to the same value.  Once n documents have been
// returned by Next, the following call will return ErrNotFound.
func (q *Query) Limit(n int) *Query {
	q.m.Lock()
	switch {
	case n == 1:
		q.limit = 1
		q.op.limit = -1
	case n == math.MinInt32: // -MinInt32 == -MinInt32
		q.limit = math.MaxInt32
		q.op.limit = math.MinInt32 + 1
	case n < 0:
		q.limit = int32(-n)
		q.op.limit = int32(n)
	default:
		q.limit = int32(n)
		q.op.limit = int32(n)
	}
	q.m.Unlock()
	return q
}

// Select enables selecting which fields should be retrieved for the results
// found. For example, the following query would only retrieve the name field:
//
//     err := collection.Find(nil).Select(bson.M{"name": 1}).One(&result)
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Retrieving+a+Subset+of+Fields
//
func (q *Query) Select(selector interface{}) *Query {
	q.m.Lock()
	q.op.selector = selector
	q.m.Unlock()
	return q
}

// Sort asks the database to order returned documents according to the
// provided field names. A field name may be prefixed by - (minus) for
// it to be sorted in reverse order.
//
// For example:
//
//     query1 := collection.Find(nil).Sort("firstname", "lastname")
//     query2 := collection.Find(nil).Sort("-age")
//     query3 := collection.Find(nil).Sort("$natural")
//     query4 := collection.Find(nil).Select(bson.M{"score": bson.M{"$meta": "textScore"}}).Sort("$textScore:score")
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Sorting+and+Natural+Order
//
func (q *Query) Sort(fields ...string) *Query {
	q.m.Lock()
	var order bson.D
	for _, field := range fields {
		n := 1
		var kind string
		if field != "" {
			if field[0] == '$' {
				if c := strings.Index(field, ":"); c > 1 && c < len(field)-1 {
					kind = field[1:c]
					field = field[c+1:]
				}
			}
			switch field[0] {
			case '+':
				field = field[1:]
			case '-':
				n = -1
				field = field[1:]
			}
		}
		if field == "" {
			panic("Sort: empty field name")
		}
		if kind == "textScore" {
			order = append(order, bson.DocElem{field, bson.M{"$meta": kind}})
		} else {
			order = append(order, bson.DocElem{field, n})
		}
	}
	q.op.options.OrderBy = order
	q.op.hasOptions = true
	q.m.Unlock()
	return q
}

// Explain returns a number of details about how the MongoDB server would
// execute the requested query, such as the number of objects examined,
// the number of times the read lock was yielded to allow writes to go in,
// and so on.
//
// For example:
//
//     m := bson.M{}
//     err := collection.Find(bson.M{"filename": name}).Explain(m)
//     if err == nil {
//         fmt.Printf("Explain: %#v\n", m)
//     }
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Optimization
//     http://www.mongodb.org/display/DOCS/Query+Optimizer
//
func (q *Query) Explain(result interface{}) error {
	q.m.Lock()
	clone := &Query{session: q.session, query: q.query}
	q.m.Unlock()
	clone.op.options.Explain = true
	clone.op.hasOptions = true
	if clone.op.limit > 0 {
		clone.op.limit = -q.op.limit
	}
	iter := clone.Iter()
	if iter.Next(result) {
		return nil
	}
	return iter.Close()
}

// TODO: Add Collection.Explain. See https://goo.gl/1MDlvz.

// Hint will include an explicit "hint" in the query to force the server
// to use a specified index, potentially improving performance in some
// situations.  The provided parameters are the fields that compose the
// key of the index to be used.  For details on how the indexKey may be
// built, see the EnsureIndex method.
//
// For example:
//
//     query := collection.Find(bson.M{"firstname": "Joe", "lastname": "Winter"})
//     query.Hint("lastname", "firstname")
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Optimization
//     http://www.mongodb.org/display/DOCS/Query+Optimizer
//
func (q *Query) Hint(indexKey ...string) *Query {
	q.m.Lock()
	keyInfo, err := parseIndexKey(indexKey)
	q.op.options.Hint = keyInfo.key
	q.op.hasOptions = true
	q.m.Unlock()
	if err != nil {
		panic(err)
	}
	return q
}

// SetMaxScan constrains the query to stop after scanning the specified
// number of documents.
//
// This modifier is generally used to prevent potentially long running
// queries from disrupting performance by scanning through too much data.
func (q *Query) SetMaxScan(n int) *Query {
	q.m.Lock()
	q.op.options.MaxScan = n
	q.op.hasOptions = true
	q.m.Unlock()
	return q
}

// SetMaxTime constrains the query to stop after running for the specified time.
//
// When the time limit is reached MongoDB automatically cancels the query.
// This can be used to efficiently prevent and identify unexpectedly slow queries.
//
// A few important notes about the mechanism enforcing this limit:
//
//  - Requests can block behind locking operations on the server, and that blocking
//    time is not accounted for. In other words, the timer starts ticking only after
//    the actual start of the query when it initially acquires the appropriate lock;
//
//  - Operations are interrupted only at interrupt points where an operation can be
//    safely aborted – the total execution time may exceed the specified value;
//
//  - The limit can be applied to both CRUD operations and commands, but not all
//    commands are interruptible;
//
//  - While iterating over results, computing follow up batches is included in the
//    total time and the iteration continues until the alloted time is over, but
//    network roundtrips are not taken into account for the limit.
//
//  - This limit does not override the inactive cursor timeout for idle cursors
//    (default is 10 min).
//
// This mechanism was introduced in MongoDB 2.6.
//
// Relevant documentation:
//
//   http://blog.mongodb.org/post/83621787773/maxtimems-and-query-optimizer-introspection-in
//
func (q *Query) SetMaxTime(d time.Duration) *Query {
	q.m.Lock()
	q.op.options.MaxTimeMS = int(d / time.Millisecond)
	q.op.hasOptions = true
	q.m.Unlock()
	return q
}

// Snapshot will force the performed query to make use of an available
// index on the _id field to prevent the same document from being returned
// more than once in a single iteration. This might happen without this
// setting in situations when the document changes in size and thus has to
// be moved while the iteration is running.
//
// Because snapshot mode traverses the _id index, it may not be used with
// sorting or explicit hints. It also cannot use any other index for the
// query.
//
// Even with snapshot mode, items inserted or deleted during the query may
// or may not be returned; that is, this mode is not a true point-in-time
// snapshot.
//
// The same effect of Snapshot may be obtained by using any unique index on
// field(s) that will not be modified (best to use Hint explicitly too).
// A non-unique index (such as creation time) may be made unique by
// appending _id to the index when creating it.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/How+to+do+Snapshotted+Queries+in+the+Mongo+Database
//
func (q *Query) Snapshot() *Query {
	q.m.Lock()
	q.op.options.Snapshot = true
	q.op.hasOptions = true
	q.m.Unlock()
	return q
}

// Comment adds a comment to the query to identify it in the database profiler output.
//
// Relevant documentation:
//
//     http://docs.mongodb.org/manual/reference/operator/meta/comment
//     http://docs.mongodb.org/manual/reference/command/profile
//     http://docs.mongodb.org/manual/administration/analyzing-mongodb-performance/#database-profiling
//
func (q *Query) Comment(comment string) *Query {
	q.m.Lock()
	q.op.options.Comment = comment
	q.op.hasOptions = true
	q.m.Unlock()
	return q
}

// LogReplay enables an option that optimizes queries that are typically
// made on the MongoDB oplog for replaying it. This is an internal
// implementation aspect and most likely uninteresting for other uses.
// It has seen at least one use case, though, so it's exposed via the API.
func (q *Query) LogReplay() *Query {
	q.m.Lock()
	q.op.flags |= flagLogReplay
	q.m.Unlock()
	return q
}

func checkQueryError(fullname string, d []byte) error {
	l := len(d)
	if l < 16 {
		return nil
	}
	if d[5] == '$' && d[6] == 'e' && d[7] == 'r' && d[8] == 'r' && d[9] == '\x00' && d[4] == '\x02' {
		goto Error
	}
	if len(fullname) < 5 || fullname[len(fullname)-5:] != ".$cmd" {
		return nil
	}
	for i := 0; i+8 < l; i++ {
		if d[i] == '\x02' && d[i+1] == 'e' && d[i+2] == 'r' && d[i+3] == 'r' && d[i+4] == 'm' && d[i+5] == 's' && d[i+6] == 'g' && d[i+7] == '\x00' {
			goto Error
		}
	}
	return nil

Error:
	result := &queryError{}
	bson.Unmarshal(d, result)
	if result.LastError != nil {
		return result.LastError
	}
	if result.Err == "" && result.ErrMsg == "" {
		return nil
	}
	if result.AssertionCode != 0 && result.Assertion != "" {
		return &QueryError{Code: result.AssertionCode, Message: result.Assertion, Assertion: true}
	}
	if result.Err != "" {
		return &QueryError{Code: result.Code, Message: result.Err}
	}
	return &QueryError{Code: result.Code, Message: result.ErrMsg}
}

// One executes the query and unmarshals the first obtained document into the
// result argument.  The result must be a struct or map value capable of being
// unmarshalled into by gobson.  This function blocks until either a result
// is available or an error happens.  For example:
//
//     err := collection.Find(bson.M{"a": 1}).One(&result)
//
// In case the resulting document includes a field named $err or errmsg, which
// are standard ways for MongoDB to return query errors, the returned err will
// be set to a *QueryError value including the Err message and the Code.  In
// those cases, the result argument is still unmarshalled into with the
// received document so that any other custom values may be obtained if
// desired.
//
func (q *Query) One(result interface{}) (err error) {
	q.m.Lock()
	session := q.session
	op := q.op // Copy.
	q.m.Unlock()

	socket, err := session.acquireSocket(true)
	if err != nil {
		return err
	}
	defer socket.Release()

	op.limit = -1

	session.prepareQuery(&op)

	expectFindReply := prepareFindOp(socket, &op, 1)

	data, err := socket.SimpleQuery(&op)
	if err != nil {
		return err
	}
	if data == nil {
		return ErrNotFound
	}
	if expectFindReply {
		var findReply struct {
			Ok     bool
			Code   int
			Errmsg string
			Cursor cursorData
		}
		err = bson.Unmarshal(data, &findReply)
		if err != nil {
			return err
		}
		if !findReply.Ok && findReply.Errmsg != "" {
			return &QueryError{Code: findReply.Code, Message: findReply.Errmsg}
		}
		if len(findReply.Cursor.FirstBatch) == 0 {
			return ErrNotFound
		}
		data = findReply.Cursor.FirstBatch[0].Data
	}
	if result != nil {
		err = bson.Unmarshal(data, result)
		if err == nil {
			debugf("Query %p document unmarshaled: %#v", q, result)
		} else {
			debugf("Query %p document unmarshaling failed: %#v", q, err)
			return err
		}
	}
	return checkQueryError(op.collection, data)
}

// prepareFindOp translates op from being an old-style wire protocol query into
// a new-style find command if that's supported by the MongoDB server (3.2+).
// It returns whether to expect a find command result or not. Note op may be
// translated into an explain command, in which case the function returns false.
func prepareFindOp(socket *mongoSocket, op *queryOp, limit int32) bool {
	if socket.ServerInfo().MaxWireVersion < 4 || op.collection == "admin.$cmd" {
		return false
	}

	nameDot := strings.Index(op.collection, ".")
	if nameDot < 0 {
		panic("invalid query collection name: " + op.collection)
	}

	find := findCmd{
		Collection:  op.collection[nameDot+1:],
		Filter:      op.query,
		Projection:  op.selector,
		Sort:        op.options.OrderBy,
		Skip:        op.skip,
		Limit:       limit,
		MaxTimeMS:   op.options.MaxTimeMS,
		MaxScan:     op.options.MaxScan,
		Hint:        op.options.Hint,
		Comment:     op.options.Comment,
		Snapshot:    op.options.Snapshot,
		OplogReplay: op.flags&flagLogReplay != 0,
	}
	if op.limit < 0 {
		find.BatchSize = -op.limit
		find.SingleBatch = true
	} else {
		find.BatchSize = op.limit
	}

	explain := op.options.Explain

	op.collection = op.collection[:nameDot] + ".$cmd"
	op.query = &find
	op.skip = 0
	op.limit = -1
	op.options = queryWrapper{}
	op.hasOptions = false

	if explain {
		op.query = bson.D{{"explain", op.query}}
		return false
	}
	return true
}

type cursorData struct {
	FirstBatch []bson.Raw "firstBatch"
	NextBatch  []bson.Raw "nextBatch"
	NS         string
	Id         int64
}

// findCmd holds the command used for performing queries on MongoDB 3.2+.
//
// Relevant documentation:
//
//     https://docs.mongodb.org/master/reference/command/find/#dbcmd.find
//
type findCmd struct {
	Collection          string      `bson:"find"`
	Filter              interface{} `bson:"filter,omitempty"`
	Sort                interface{} `bson:"sort,omitempty"`
	Projection          interface{} `bson:"projection,omitempty"`
	Hint                interface{} `bson:"hint,omitempty"`
	Skip                interface{} `bson:"skip,omitempty"`
	Limit               int32       `bson:"limit,omitempty"`
	BatchSize           int32       `bson:"batchSize,omitempty"`
	SingleBatch         bool        `bson:"singleBatch,omitempty"`
	Comment             string      `bson:"comment,omitempty"`
	MaxScan             int         `bson:"maxScan,omitempty"`
	MaxTimeMS           int         `bson:"maxTimeMS,omitempty"`
	ReadConcern         interface{} `bson:"readConcern,omitempty"`
	Max                 interface{} `bson:"max,omitempty"`
	Min                 interface{} `bson:"min,omitempty"`
	ReturnKey           bool        `bson:"returnKey,omitempty"`
	ShowRecordId        bool        `bson:"showRecordId,omitempty"`
	Snapshot            bool        `bson:"snapshot,omitempty"`
	Tailable            bool        `bson:"tailable,omitempty"`
	AwaitData           bool        `bson:"awaitData,omitempty"`
	OplogReplay         bool        `bson:"oplogReplay,omitempty"`
	NoCursorTimeout     bool        `bson:"noCursorTimeout,omitempty"`
	AllowPartialResults bool        `bson:"allowPartialResults,omitempty"`
}

// getMoreCmd holds the command used for requesting more query results on MongoDB 3.2+.
//
// Relevant documentation:
//
//     https://docs.mongodb.org/master/reference/command/getMore/#dbcmd.getMore
//
type getMoreCmd struct {
	CursorId   int64  `bson:"getMore"`
	Collection string `bson:"collection"`
	BatchSize  int32  `bson:"batchSize,omitempty"`
	MaxTimeMS  int64  `bson:"maxTimeMS,omitempty"`
}

// run duplicates the behavior of collection.Find(query).One(&result)
// as performed by Database.Run, specializing the logic for running
// database commands on a given socket.
func (db *Database) run(socket *mongoSocket, cmd, result interface{}) (err error) {
	// Database.Run:
	if name, ok := cmd.(string); ok {
		cmd = bson.D{{name, 1}}
	}

	// Collection.Find:
	session := db.Session
	session.m.RLock()
	op := session.queryConfig.op // Copy.
	session.m.RUnlock()
	op.query = cmd
	op.collection = db.Name + ".$cmd"

	// Query.One:
	session.prepareQuery(&op)
	op.limit = -1

	data, err := socket.SimpleQuery(&op)
	if err != nil {
		return err
	}
	if data == nil {
		return ErrNotFound
	}
	if result != nil {
		err = bson.Unmarshal(data, result)
		if err == nil {
			var res bson.M
			bson.Unmarshal(data, &res)
			debugf("Run command unmarshaled: %#v, result: %#v", op, res)
		} else {
			debugf("Run command unmarshaling failed: %#v", op, err)
			return err
		}
	}
	return checkQueryError(op.collection, data)
}

// The DBRef type implements support for the database reference MongoDB
// convention as supported by multiple drivers.  This convention enables
// cross-referencing documents between collections and databases using
// a structure which includes a collection name, a document id, and
// optionally a database name.
//
// See the FindRef methods on Session and on Database.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Database+References
//
type DBRef struct {
	Collection string      `bson:"$ref"`
	Id         interface{} `bson:"$id"`
	Database   string      `bson:"$db,omitempty"`
}

// NOTE: Order of fields for DBRef above does matter, per documentation.

// FindRef returns a query that looks for the document in the provided
// reference. If the reference includes the DB field, the document will
// be retrieved from the respective database.
//
// See also the DBRef type and the FindRef method on Session.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Database+References
//
func (db *Database) FindRef(ref *DBRef) *Query {
	var c *Collection
	if ref.Database == "" {
		c = db.C(ref.Collection)
	} else {
		c = db.Session.DB(ref.Database).C(ref.Collection)
	}
	return c.FindId(ref.Id)
}

// FindRef returns a query that looks for the document in the provided
// reference. For a DBRef to be resolved correctly at the session level
// it must necessarily have the optional DB field defined.
//
// See also the DBRef type and the FindRef method on Database.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Database+References
//
func (s *Session) FindRef(ref *DBRef) *Query {
	if ref.Database == "" {
		panic(errors.New(fmt.Sprintf("Can't resolve database for %#v", ref)))
	}
	c := s.DB(ref.Database).C(ref.Collection)
	return c.FindId(ref.Id)
}

// CollectionNames returns the collection names present in the db database.
func (db *Database) CollectionNames() (names []string, err error) {
	// Clone session and set it to Monotonic mode so that the server
	// used for the query may be safely obtained afterwards, if
	// necessary for iteration when a cursor is received.
	cloned := db.Session.nonEventual()
	defer cloned.Close()

	batchSize := int(cloned.queryConfig.op.limit)

	// Try with a command.
	var result struct {
		Collections []bson.Raw
		Cursor      cursorData
	}
	err = db.With(cloned).Run(bson.D{{"listCollections", 1}, {"cursor", bson.D{{"batchSize", batchSize}}}}, &result)
	if err == nil {
		firstBatch := result.Collections
		if firstBatch == nil {
			firstBatch = result.Cursor.FirstBatch
		}
		var iter *Iter
		ns := strings.SplitN(result.Cursor.NS, ".", 2)
		if len(ns) < 2 {
			iter = db.With(cloned).C("").NewIter(nil, firstBatch, result.Cursor.Id, nil)
		} else {
			iter = cloned.DB(ns[0]).C(ns[1]).NewIter(nil, firstBatch, result.Cursor.Id, nil)
		}
		var coll struct{ Name string }
		for iter.Next(&coll) {
			names = append(names, coll.Name)
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
		sort.Strings(names)
		return names, err
	}
	if err != nil && !isNoCmd(err) {
		return nil, err
	}

	// Command not yet supported. Query the database instead.
	nameIndex := len(db.Name) + 1
	iter := db.C("system.namespaces").Find(nil).Iter()
	var coll struct{ Name string }
	for iter.Next(&coll) {
		if strings.Index(coll.Name, "$") < 0 || strings.Index(coll.Name, ".oplog.$") >= 0 {
			names = append(names, coll.Name[nameIndex:])
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

type dbNames struct {
	Databases []struct {
		Name  string
		Empty bool
	}
}

// DatabaseNames returns the names of non-empty databases present in the cluster.
func (s *Session) DatabaseNames() (names []string, err error) {
	var result dbNames
	err = s.Run("listDatabases", &result)
	if err != nil {
		return nil, err
	}
	for _, db := range result.Databases {
		if !db.Empty {
			names = append(names, db.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// Iter executes the query and returns an iterator capable of going over all
// the results. Results will be returned in batches of configurable
// size (see the Batch method) and more documents will be requested when a
// configurable number of documents is iterated over (see the Prefetch method).
func (q *Query) Iter() *Iter {
	q.m.Lock()
	session := q.session
	op := q.op
	prefetch := q.prefetch
	limit := q.limit
	q.m.Unlock()

	iter := &Iter{
		session:  session,
		prefetch: prefetch,
		limit:    limit,
		timeout:  -1,
	}
	iter.gotReply.L = &iter.m
	iter.op.collection = op.collection
	iter.op.limit = op.limit
	iter.op.replyFunc = iter.replyFunc()
	iter.docsToReceive++

	socket, err := session.acquireSocket(true)
	if err != nil {
		iter.err = err
		return iter
	}
	defer socket.Release()

	session.prepareQuery(&op)
	op.replyFunc = iter.op.replyFunc

	if prepareFindOp(socket, &op, limit) {
		iter.findCmd = true
	}

	iter.server = socket.Server()
	err = socket.Query(&op)
	if err != nil {
		// Must lock as the query is already out and it may call replyFunc.
		iter.m.Lock()
		iter.err = err
		iter.m.Unlock()
	}

	return iter
}

// Tail returns a tailable iterator. Unlike a normal iterator, a
// tailable iterator may wait for new values to be inserted in the
// collection once the end of the current result set is reached,
// A tailable iterator may only be used with capped collections.
//
// The timeout parameter indicates how long Next will block waiting
// for a result before timing out.  If set to -1, Next will not
// timeout, and will continue waiting for a result for as long as
// the cursor is valid and the session is not closed. If set to 0,
// Next times out as soon as it reaches the end of the result set.
// Otherwise, Next will wait for at least the given number of
// seconds for a new document to be available before timing out.
//
// On timeouts, Next will unblock and return false, and the Timeout
// method will return true if called. In these cases, Next may still
// be called again on the same iterator to check if a new value is
// available at the current cursor position, and again it will block
// according to the specified timeoutSecs. If the cursor becomes
// invalid, though, both Next and Timeout will return false and
// the query must be restarted.
//
// The following example demonstrates timeout handling and query
// restarting:
//
//    iter := collection.Find(nil).Sort("$natural").Tail(5 * time.Second)
//    for {
//         for iter.Next(&result) {
//             fmt.Println(result.Id)
//             lastId = result.Id
//         }
//         if iter.Err() != nil {
//             return iter.Close()
//         }
//         if iter.Timeout() {
//             continue
//         }
//         query := collection.Find(bson.M{"_id": bson.M{"$gt": lastId}})
//         iter = query.Sort("$natural").Tail(5 * time.Second)
//    }
//    iter.Close()
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Tailable+Cursors
//     http://www.mongodb.org/display/DOCS/Capped+Collections
//     http://www.mongodb.org/display/DOCS/Sorting+and+Natural+Order
//
func (q *Query) Tail(timeout time.Duration) *Iter {
	q.m.Lock()
	session := q.session
	op := q.op
	prefetch := q.prefetch
	q.m.Unlock()

	iter := &Iter{session: session, prefetch: prefetch}
	iter.gotReply.L = &iter.m
	iter.timeout = timeout
	iter.op.collection = op.collection
	iter.op.limit = op.limit
	iter.op.replyFunc = iter.replyFunc()
	iter.docsToReceive++
	session.prepareQuery(&op)
	op.replyFunc = iter.op.replyFunc
	op.flags |= flagTailable | flagAwaitData

	socket, err := session.acquireSocket(true)
	if err != nil {
		iter.err = err
	} else {
		iter.server = socket.Server()
		err = socket.Query(&op)
		if err != nil {
			// Must lock as the query is already out and it may call replyFunc.
			iter.m.Lock()
			iter.err = err
			iter.m.Unlock()
		}
		socket.Release()
	}
	return iter
}

func (s *Session) prepareQuery(op *queryOp) {
	s.m.RLock()
	op.mode = s.consistency
	if s.slaveOk {
		op.flags |= flagSlaveOk
	}
	s.m.RUnlock()
	return
}

// Err returns nil if no errors happened during iteration, or the actual
// error otherwise.
//
// In case a resulting document included a field named $err or errmsg, which are
// standard ways for MongoDB to report an improper query, the returned value has
// a *QueryError type, and includes the Err message and the Code.
func (iter *Iter) Err() error {
	iter.m.Lock()
	err := iter.err
	iter.m.Unlock()
	if err == ErrNotFound {
		return nil
	}
	return err
}

// Close kills the server cursor used by the iterator, if any, and returns
// nil if no errors happened during iteration, or the actual error otherwise.
//
// Server cursors are automatically closed at the end of an iteration, which
// means close will do nothing unless the iteration was interrupted before
// the server finished sending results to the driver. If Close is not called
// in such a situation, the cursor will remain available at the server until
// the default cursor timeout period is reached. No further problems arise.
//
// Close is idempotent. That means it can be called repeatedly and will
// return the same result every time.
//
// In case a resulting document included a field named $err or errmsg, which are
// standard ways for MongoDB to report an improper query, the returned value has
// a *QueryError type.
func (iter *Iter) Close() error {
	iter.m.Lock()
	cursorId := iter.op.cursorId
	iter.op.cursorId = 0
	err := iter.err
	iter.m.Unlock()
	if cursorId == 0 {
		if err == ErrNotFound {
			return nil
		}
		return err
	}
	socket, err := iter.acquireSocket()
	if err == nil {
		// TODO Batch kills.
		err = socket.Query(&killCursorsOp{[]int64{cursorId}})
		socket.Release()
	}

	iter.m.Lock()
	if err != nil && (iter.err == nil || iter.err == ErrNotFound) {
		iter.err = err
	} else if iter.err != ErrNotFound {
		err = iter.err
	}
	iter.m.Unlock()
	return err
}

// Timeout returns true if Next returned false due to a timeout of
// a tailable cursor. In those cases, Next may be called again to continue
// the iteration at the previous cursor position.
func (iter *Iter) Timeout() bool {
	iter.m.Lock()
	result := iter.timedout
	iter.m.Unlock()
	return result
}

// Next retrieves the next document from the result set, blocking if necessary.
// This method will also automatically retrieve another batch of documents from
// the server when the current one is exhausted, or before that in background
// if pre-fetching is enabled (see the Query.Prefetch and Session.SetPrefetch
// methods).
//
// Next returns true if a document was successfully unmarshalled onto result,
// and false at the end of the result set or if an error happened.
// When Next returns false, the Err method should be called to verify if
// there was an error during iteration.
//
// For example:
//
//    iter := collection.Find(nil).Iter()
//    for iter.Next(&result) {
//        fmt.Printf("Result: %v\n", result.Id)
//    }
//    if err := iter.Close(); err != nil {
//        return err
//    }
//
func (iter *Iter) Next(result interface{}) bool {
	iter.m.Lock()
	iter.timedout = false
	timeout := time.Time{}
	for iter.err == nil && iter.docData.Len() == 0 && (iter.docsToReceive > 0 || iter.op.cursorId != 0) {
		if iter.docsToReceive == 0 {
			if iter.timeout >= 0 {
				if timeout.IsZero() {
					timeout = time.Now().Add(iter.timeout)
				}
				if time.Now().After(timeout) {
					iter.timedout = true
					iter.m.Unlock()
					return false
				}
			}
			iter.getMore()
			if iter.err != nil {
				break
			}
		}
		iter.gotReply.Wait()
	}

	// Exhaust available data before reporting any errors.
	if docData, ok := iter.docData.Pop().([]byte); ok {
		close := false
		if iter.limit > 0 {
			iter.limit--
			if iter.limit == 0 {
				if iter.docData.Len() > 0 {
					iter.m.Unlock()
					panic(fmt.Errorf("data remains after limit exhausted: %d", iter.docData.Len()))
				}
				iter.err = ErrNotFound
				close = true
			}
		}
		if iter.op.cursorId != 0 && iter.err == nil {
			iter.docsBeforeMore--
			if iter.docsBeforeMore == -1 {
				iter.getMore()
			}
		}
		iter.m.Unlock()

		if close {
			iter.Close()
		}
		err := bson.Unmarshal(docData, result)
		if err != nil {
			debugf("Iter %p document unmarshaling failed: %#v", iter, err)
			iter.m.Lock()
			if iter.err == nil {
				iter.err = err
			}
			iter.m.Unlock()
			return false
		}
		debugf("Iter %p document unmarshaled: %#v", iter, result)
		// XXX Only have to check first document for a query error?
		err = checkQueryError(iter.op.collection, docData)
		if err != nil {
			iter.m.Lock()
			if iter.err == nil {
				iter.err = err
			}
			iter.m.Unlock()
			return false
		}
		return true
	} else if iter.err != nil {
		debugf("Iter %p returning false: %s", iter, iter.err)
		iter.m.Unlock()
		return false
	} else if iter.op.cursorId == 0 {
		iter.err = ErrNotFound
		debugf("Iter %p exhausted with cursor=0", iter)
		iter.m.Unlock()
		return false
	}

	panic("unreachable")
}

// All retrieves all documents from the result set into the provided slice
// and closes the iterator.
//
// The result argument must necessarily be the address for a slice. The slice
// may be nil or previously allocated.
//
// WARNING: Obviously, All must not be used with result sets that may be
// potentially large, since it may consume all memory until the system
// crashes. Consider building the query with a Limit clause to ensure the
// result size is bounded.
//
// For instance:
//
//    var result []struct{ Value int }
//    iter := collection.Find(nil).Limit(100).Iter()
//    err := iter.All(&result)
//    if err != nil {
//        return err
//    }
//
func (iter *Iter) All(result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}
	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			if !iter.Next(elemp.Interface()) {
				break
			}
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			if !iter.Next(slicev.Index(i).Addr().Interface()) {
				break
			}
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return iter.Close()
}

// All works like Iter.All.
func (q *Query) All(result interface{}) error {
	return q.Iter().All(result)
}

// The For method is obsolete and will be removed in a future release.
// See Iter as an elegant replacement.
func (q *Query) For(result interface{}, f func() error) error {
	return q.Iter().For(result, f)
}

// The For method is obsolete and will be removed in a future release.
// See Iter as an elegant replacement.
func (iter *Iter) For(result interface{}, f func() error) (err error) {
	valid := false
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		switch v.Kind() {
		case reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
			valid = v.IsNil()
		}
	}
	if !valid {
		panic("For needs a pointer to nil reference value.  See the documentation.")
	}
	zero := reflect.Zero(v.Type())
	for {
		v.Set(zero)
		if !iter.Next(result) {
			break
		}
		err = f()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}

// acquireSocket acquires a socket from the same server that the iterator
// cursor was obtained from.
//
// WARNING: This method must not be called with iter.m locked. Acquiring the
// socket depends on the cluster sync loop, and the cluster sync loop might
// attempt actions which cause replyFunc to be called, inducing a deadlock.
func (iter *Iter) acquireSocket() (*mongoSocket, error) {
	socket, err := iter.session.acquireSocket(true)
	if err != nil {
		return nil, err
	}
	if socket.Server() != iter.server {
		// Socket server changed during iteration. This may happen
		// with Eventual sessions, if a Refresh is done, or if a
		// monotonic session gets a write and shifts from secondary
		// to primary. Our cursor is in a specific server, though.
		iter.session.m.Lock()
		sockTimeout := iter.session.sockTimeout
		iter.session.m.Unlock()
		socket.Release()
		socket, _, err = iter.server.AcquireSocket(0, sockTimeout)
		if err != nil {
			return nil, err
		}
		err := iter.session.socketLogin(socket)
		if err != nil {
			socket.Release()
			return nil, err
		}
	}
	return socket, nil
}

func (iter *Iter) getMore() {
	// Increment now so that unlocking the iterator won't cause a
	// different goroutine to get here as well.
	iter.docsToReceive++
	iter.m.Unlock()
	socket, err := iter.acquireSocket()
	iter.m.Lock()
	if err != nil {
		iter.err = err
		return
	}
	defer socket.Release()

	debugf("Iter %p requesting more documents", iter)
	if iter.limit > 0 {
		// The -1 below accounts for the fact docsToReceive was incremented above.
		limit := iter.limit - int32(iter.docsToReceive-1) - int32(iter.docData.Len())
		if limit < iter.op.limit {
			iter.op.limit = limit
		}
	}
	var op interface{}
	if iter.findCmd {
		op = iter.getMoreCmd()
	} else {
		op = &iter.op
	}
	if err := socket.Query(op); err != nil {
		iter.docsToReceive--
		iter.err = err
	}
}

func (iter *Iter) getMoreCmd() *queryOp {
	// TODO: Define the query statically in the Iter type, next to getMoreOp.
	nameDot := strings.Index(iter.op.collection, ".")
	if nameDot < 0 {
		panic("invalid query collection name: " + iter.op.collection)
	}

	getMore := getMoreCmd{
		CursorId:   iter.op.cursorId,
		Collection: iter.op.collection[nameDot+1:],
		BatchSize:  iter.op.limit,
	}

	var op queryOp
	op.collection = iter.op.collection[:nameDot] + ".$cmd"
	op.query = &getMore
	op.limit = -1
	op.replyFunc = iter.op.replyFunc
	return &op
}

type countCmd struct {
	Count string
	Query interface{}
	Limit int32 ",omitempty"
	Skip  int32 ",omitempty"
}

// Count returns the total number of documents in the result set.
func (q *Query) Count() (n int, err error) {
	q.m.Lock()
	session := q.session
	op := q.op
	limit := q.limit
	q.m.Unlock()

	c := strings.Index(op.collection, ".")
	if c < 0 {
		return 0, errors.New("Bad collection name: " + op.collection)
	}

	dbname := op.collection[:c]
	cname := op.collection[c+1:]
	query := op.query
	if query == nil {
		query = bson.D{}
	}
	result := struct{ N int }{}
	err = session.DB(dbname).Run(countCmd{cname, query, limit, op.skip}, &result)
	return result.N, err
}

// Count returns the total number of documents in the collection.
func (c *Collection) Count() (n int, err error) {
	return c.Find(nil).Count()
}

type distinctCmd struct {
	Collection string "distinct"
	Key        string
	Query      interface{} ",omitempty"
}

// Distinct unmarshals into result the list of distinct values for the given key.
//
// For example:
//
//     var result []int
//     err := collection.Find(bson.M{"gender": "F"}).Distinct("age", &result)
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/Aggregation
//
func (q *Query) Distinct(key string, result interface{}) error {
	q.m.Lock()
	session := q.session
	op := q.op // Copy.
	q.m.Unlock()

	c := strings.Index(op.collection, ".")
	if c < 0 {
		return errors.New("Bad collection name: " + op.collection)
	}

	dbname := op.collection[:c]
	cname := op.collection[c+1:]

	var doc struct{ Values bson.Raw }
	err := session.DB(dbname).Run(distinctCmd{cname, key, op.query}, &doc)
	if err != nil {
		return err
	}
	return doc.Values.Unmarshal(result)
}

type mapReduceCmd struct {
	Collection string "mapreduce"
	Map        string ",omitempty"
	Reduce     string ",omitempty"
	Finalize   string ",omitempty"
	Limit      int32  ",omitempty"
	Out        interface{}
	Query      interface{} ",omitempty"
	Sort       interface{} ",omitempty"
	Scope      interface{} ",omitempty"
	Verbose    bool        ",omitempty"
}

type mapReduceResult struct {
	Results    bson.Raw
	Result     bson.Raw
	TimeMillis int64 "timeMillis"
	Counts     struct{ Input, Emit, Output int }
	Ok         bool
	Err        string
	Timing     *MapReduceTime
}

type MapReduce struct {
	Map      string      // Map Javascript function code (required)
	Reduce   string      // Reduce Javascript function code (required)
	Finalize string      // Finalize Javascript function code (optional)
	Out      interface{} // Output collection name or document. If nil, results are inlined into the result parameter.
	Scope    interface{} // Optional global scope for Javascript functions
	Verbose  bool
}

type MapReduceInfo struct {
	InputCount  int            // Number of documents mapped
	EmitCount   int            // Number of times reduce called emit
	OutputCount int            // Number of documents in resulting collection
	Database    string         // Output database, if results are not inlined
	Collection  string         // Output collection, if results are not inlined
	Time        int64          // Time to run the job, in nanoseconds
	VerboseTime *MapReduceTime // Only defined if Verbose was true
}

type MapReduceTime struct {
	Total    int64 // Total time, in nanoseconds
	Map      int64 "mapTime"  // Time within map function, in nanoseconds
	EmitLoop int64 "emitLoop" // Time within the emit/map loop, in nanoseconds
}

// MapReduce executes a map/reduce job for documents covered by the query.
// That kind of job is suitable for very flexible bulk aggregation of data
// performed at the server side via Javascript functions.
//
// Results from the job may be returned as a result of the query itself
// through the result parameter in case they'll certainly fit in memory
// and in a single document.  If there's the possibility that the amount
// of data might be too large, results must be stored back in an alternative
// collection or even a separate database, by setting the Out field of the
// provided MapReduce job.  In that case, provide nil as the result parameter.
//
// These are some of the ways to set Out:
//
//     nil
//         Inline results into the result parameter.
//
//     bson.M{"replace": "mycollection"}
//         The output will be inserted into a collection which replaces any
//         existing collection with the same name.
//
//     bson.M{"merge": "mycollection"}
//         This option will merge new data into the old output collection. In
//         other words, if the same key exists in both the result set and the
//         old collection, the new key will overwrite the old one.
//
//     bson.M{"reduce": "mycollection"}
//         If documents exist for a given key in the result set and in the old
//         collection, then a reduce operation (using the specified reduce
//         function) will be performed on the two values and the result will be
//         written to the output collection. If a finalize function was
//         provided, this will be run after the reduce as well.
//
//     bson.M{...., "db": "mydb"}
//         Any of the above options can have the "db" key included for doing
//         the respective action in a separate database.
//
// The following is a trivial example which will count the number of
// occurrences of a field named n on each document in a collection, and
// will return results inline:
//
//     job := &mgo.MapReduce{
//             Map:      "function() { emit(this.n, 1) }",
//             Reduce:   "function(key, values) { return Array.sum(values) }",
//     }
//     var result []struct { Id int "_id"; Value int }
//     _, err := collection.Find(nil).MapReduce(job, &result)
//     if err != nil {
//         return err
//     }
//     for _, item := range result {
//         fmt.Println(item.Value)
//     }
//
// This function is compatible with MongoDB 1.7.4+.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/MapReduce
//
func (q *Query) MapReduce(job *MapReduce, result interface{}) (info *MapReduceInfo, err error) {
	q.m.Lock()
	session := q.session
	op := q.op // Copy.
	limit := q.limit
	q.m.Unlock()

	c := strings.Index(op.collection, ".")
	if c < 0 {
		return nil, errors.New("Bad collection name: " + op.collection)
	}

	dbname := op.collection[:c]
	cname := op.collection[c+1:]

	cmd := mapReduceCmd{
		Collection: cname,
		Map:        job.Map,
		Reduce:     job.Reduce,
		Finalize:   job.Finalize,
		Out:        fixMROut(job.Out),
		Scope:      job.Scope,
		Verbose:    job.Verbose,
		Query:      op.query,
		Sort:       op.options.OrderBy,
		Limit:      limit,
	}

	if cmd.Out == nil {
		cmd.Out = bson.D{{"inline", 1}}
	}

	var doc mapReduceResult
	err = session.DB(dbname).Run(&cmd, &doc)
	if err != nil {
		return nil, err
	}
	if doc.Err != "" {
		return nil, errors.New(doc.Err)
	}

	info = &MapReduceInfo{
		InputCount:  doc.Counts.Input,
		EmitCount:   doc.Counts.Emit,
		OutputCount: doc.Counts.Output,
		Time:        doc.TimeMillis * 1e6,
	}

	if doc.Result.Kind == 0x02 {
		err = doc.Result.Unmarshal(&info.Collection)
		info.Database = dbname
	} else if doc.Result.Kind == 0x03 {
		var v struct{ Collection, Db string }
		err = doc.Result.Unmarshal(&v)
		info.Collection = v.Collection
		info.Database = v.Db
	}

	if doc.Timing != nil {
		info.VerboseTime = doc.Timing
		info.VerboseTime.Total *= 1e6
		info.VerboseTime.Map *= 1e6
		info.VerboseTime.EmitLoop *= 1e6
	}

	if err != nil {
		return nil, err
	}
	if result != nil {
		return info, doc.Results.Unmarshal(result)
	}
	return info, nil
}

// The "out" option in the MapReduce command must be ordered. This was
// found after the implementation was accepting maps for a long time,
// so rather than breaking the API, we'll fix the order if necessary.
// Details about the order requirement may be seen in MongoDB's code:
//
//     http://goo.gl/L8jwJX
//
func fixMROut(out interface{}) interface{} {
	outv := reflect.ValueOf(out)
	if outv.Kind() != reflect.Map || outv.Type().Key() != reflect.TypeOf("") {
		return out
	}
	outs := make(bson.D, outv.Len())

	outTypeIndex := -1
	for i, k := range outv.MapKeys() {
		ks := k.String()
		outs[i].Name = ks
		outs[i].Value = outv.MapIndex(k).Interface()
		switch ks {
		case "normal", "replace", "merge", "reduce", "inline":
			outTypeIndex = i
		}
	}
	if outTypeIndex > 0 {
		outs[0], outs[outTypeIndex] = outs[outTypeIndex], outs[0]
	}
	return outs
}

// Change holds fields for running a findAndModify MongoDB command via
// the Query.Apply method.
type Change struct {
	Update    interface{} // The update document
	Upsert    bool        // Whether to insert in case the document isn't found
	Remove    bool        // Whether to remove the document found rather than updating
	ReturnNew bool        // Should the modified document be returned rather than the old one
}

type findModifyCmd struct {
	Collection                  string      "findAndModify"
	Query, Update, Sort, Fields interface{} ",omitempty"
	Upsert, Remove, New         bool        ",omitempty"
}

type valueResult struct {
	Value     bson.Raw
	LastError LastError "lastErrorObject"
}

// Apply runs the findAndModify MongoDB command, which allows updating, upserting
// or removing a document matching a query and atomically returning either the old
// version (the default) or the new version of the document (when ReturnNew is true).
// If no objects are found Apply returns ErrNotFound.
//
// The Sort and Select query methods affect the result of Apply.  In case
// multiple documents match the query, Sort enables selecting which document to
// act upon by ordering it first.  Select enables retrieving only a selection
// of fields of the new or old document.
//
// This simple example increments a counter and prints its new value:
//
//     change := mgo.Change{
//             Update: bson.M{"$inc": bson.M{"n": 1}},
//             ReturnNew: true,
//     }
//     info, err = col.Find(M{"_id": id}).Apply(change, &doc)
//     fmt.Println(doc.N)
//
// This method depends on MongoDB >= 2.0 to work properly.
//
// Relevant documentation:
//
//     http://www.mongodb.org/display/DOCS/findAndModify+Command
//     http://www.mongodb.org/display/DOCS/Updating
//     http://www.mongodb.org/display/DOCS/Atomic+Operations
//
func (q *Query) Apply(change Change, result interface{}) (info *ChangeInfo, err error) {
	q.m.Lock()
	session := q.session
	op := q.op // Copy.
	q.m.Unlock()

	c := strings.Index(op.collection, ".")
	if c < 0 {
		return nil, errors.New("bad collection name: " + op.collection)
	}

	dbname := op.collection[:c]
	cname := op.collection[c+1:]

	cmd := findModifyCmd{
		Collection: cname,
		Update:     change.Update,
		Upsert:     change.Upsert,
		Remove:     change.Remove,
		New:        change.ReturnNew,
		Query:      op.query,
		Sort:       op.options.OrderBy,
		Fields:     op.selector,
	}

	session = session.Clone()
	defer session.Close()
	session.SetMode(Strong, false)

	var doc valueResult
	err = session.DB(dbname).Run(&cmd, &doc)
	if err != nil {
		if qerr, ok := err.(*QueryError); ok && qerr.Message == "No matching object found" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if doc.LastError.N == 0 {
		return nil, ErrNotFound
	}
	if doc.Value.Kind != 0x0A && result != nil {
		err = doc.Value.Unmarshal(result)
		if err != nil {
			return nil, err
		}
	}
	info = &ChangeInfo{}
	lerr := &doc.LastError
	if lerr.UpdatedExisting {
		info.Updated = lerr.N
		info.Matched = lerr.N
	} else if change.Remove {
		info.Removed = lerr.N
		info.Matched = lerr.N
	} else if change.Upsert {
		info.UpsertedId = lerr.UpsertedId
	}
	return info, nil
}

// The BuildInfo type encapsulates details about the running MongoDB server.
//
// Note that the VersionArray field was introduced in MongoDB 2.0+, but it is
// internally assembled from the Version information for previous versions.
// In both cases, VersionArray is guaranteed to have at least 4 entries.
type BuildInfo struct {
	Version        string
	VersionArray   []int  `bson:"versionArray"` // On MongoDB 2.0+; assembled from Version otherwise
	GitVersion     string `bson:"gitVersion"`
	OpenSSLVersion string `bson:"OpenSSLVersion"`
	SysInfo        string `bson:"sysInfo"` // Deprecated and empty on MongoDB 3.2+.
	Bits           int
	Debug          bool
	MaxObjectSize  int `bson:"maxBsonObjectSize"`
}

// VersionAtLeast returns whether the BuildInfo version is greater than or
// equal to the provided version number. If more than one number is
// provided, numbers will be considered as major, minor, and so on.
func (bi *BuildInfo) VersionAtLeast(version ...int) bool {
	for i := range version {
		if i == len(bi.VersionArray) {
			return false
		}
		if bi.VersionArray[i] < version[i] {
			return false
		}
	}
	return true
}

// BuildInfo retrieves the version and other details about the
// running MongoDB server.
func (s *Session) BuildInfo() (info BuildInfo, err error) {
	err = s.Run(bson.D{{"buildInfo", "1"}}, &info)
	if len(info.VersionArray) == 0 {
		for _, a := range strings.Split(info.Version, ".") {
			i, err := strconv.Atoi(a)
			if err != nil {
				break
			}
			info.VersionArray = append(info.VersionArray, i)
		}
	}
	for len(info.VersionArray) < 4 {
		info.VersionArray = append(info.VersionArray, 0)
	}
	if i := strings.IndexByte(info.GitVersion, ' '); i >= 0 {
		// Strip off the " modules: enterprise" suffix. This is a _git version_.
		// That information may be moved to another field if people need it.
		info.GitVersion = info.GitVersion[:i]
	}
	if info.SysInfo == "deprecated" {
		info.SysInfo = ""
	}
	return
}

// ---------------------------------------------------------------------------
// Internal session handling helpers.

func (s *Session) acquireSocket(slaveOk bool) (*mongoSocket, error) {

	// Read-only lock to check for previously reserved socket.
	s.m.RLock()
	// If there is a slave socket reserved and its use is acceptable, take it as long
	// as there isn't a master socket which would be preferred by the read preference mode.
	if s.slaveSocket != nil && s.slaveOk && slaveOk && (s.masterSocket == nil || s.consistency != PrimaryPreferred && s.consistency != Monotonic) {
		socket := s.slaveSocket
		socket.Acquire()
		s.m.RUnlock()
		return socket, nil
	}
	if s.masterSocket != nil {
		socket := s.masterSocket
		socket.Acquire()
		s.m.RUnlock()
		return socket, nil
	}
	s.m.RUnlock()

	// No go.  We may have to request a new socket and change the session,
	// so try again but with an exclusive lock now.
	s.m.Lock()
	defer s.m.Unlock()

	if s.slaveSocket != nil && s.slaveOk && slaveOk && (s.masterSocket == nil || s.consistency != PrimaryPreferred && s.consistency != Monotonic) {
		s.slaveSocket.Acquire()
		return s.slaveSocket, nil
	}
	if s.masterSocket != nil {
		s.masterSocket.Acquire()
		return s.masterSocket, nil
	}

	// Still not good.  We need a new socket.
	sock, err := s.cluster().AcquireSocket(s.consistency, slaveOk && s.slaveOk, s.syncTimeout, s.sockTimeout, s.queryConfig.op.serverTags, s.poolLimit)
	if err != nil {
		return nil, err
	}

	// Authenticate the new socket.
	if err = s.socketLogin(sock); err != nil {
		sock.Release()
		return nil, err
	}

	// Keep track of the new socket, if necessary.
	// Note that, as a special case, if the Eventual session was
	// not refreshed (s.slaveSocket != nil), it means the developer
	// asked to preserve an existing reserved socket, so we'll
	// keep a master one around too before a Refresh happens.
	if s.consistency != Eventual || s.slaveSocket != nil {
		s.setSocket(sock)
	}

	// Switch over a Monotonic session to the master.
	if !slaveOk && s.consistency == Monotonic {
		s.slaveOk = false
	}

	return sock, nil
}

// setSocket binds socket to this section.
func (s *Session) setSocket(socket *mongoSocket) {
	info := socket.Acquire()
	if info.Master {
		if s.masterSocket != nil {
			panic("setSocket(master) with existing master socket reserved")
		}
		s.masterSocket = socket
	} else {
		if s.slaveSocket != nil {
			panic("setSocket(slave) with existing slave socket reserved")
		}
		s.slaveSocket = socket
	}
}

// unsetSocket releases any slave and/or master sockets reserved.
func (s *Session) unsetSocket() {
	if s.masterSocket != nil {
		s.masterSocket.Release()
	}
	if s.slaveSocket != nil {
		s.slaveSocket.Release()
	}
	s.masterSocket = nil
	s.slaveSocket = nil
}

func (iter *Iter) replyFunc() replyFunc {
	return func(err error, op *replyOp, docNum int, docData []byte) {
		iter.m.Lock()
		iter.docsToReceive--
		if err != nil {
			iter.err = err
			debugf("Iter %p received an error: %s", iter, err.Error())
		} else if docNum == -1 {
			debugf("Iter %p received no documents (cursor=%d).", iter, op.cursorId)
			if op != nil && op.cursorId != 0 {
				// It's a tailable cursor.
				iter.op.cursorId = op.cursorId
			} else if op != nil && op.cursorId == 0 && op.flags&1 == 1 {
				// Cursor likely timed out.
				iter.err = ErrCursor
			} else {
				iter.err = ErrNotFound
			}
		} else if iter.findCmd {
			debugf("Iter %p received reply document %d/%d (cursor=%d)", iter, docNum+1, int(op.replyDocs), op.cursorId)
			var findReply struct {
				Ok     bool
				Code   int
				Errmsg string
				Cursor cursorData
			}
			if err := bson.Unmarshal(docData, &findReply); err != nil {
				iter.err = err
			} else if !findReply.Ok && findReply.Errmsg != "" {
				iter.err = &QueryError{Code: findReply.Code, Message: findReply.Errmsg}
			} else if len(findReply.Cursor.FirstBatch) == 0 && len(findReply.Cursor.NextBatch) == 0 {
				iter.err = ErrNotFound
			} else {
				batch := findReply.Cursor.FirstBatch
				if len(batch) == 0 {
					batch = findReply.Cursor.NextBatch
				}
				rdocs := len(batch)
				for _, raw := range batch {
					iter.docData.Push(raw.Data)
				}
				iter.docsToReceive = 0
				docsToProcess := iter.docData.Len()
				if iter.limit == 0 || int32(docsToProcess) < iter.limit {
					iter.docsBeforeMore = docsToProcess - int(iter.prefetch*float64(rdocs))
				} else {
					iter.docsBeforeMore = -1
				}
				iter.op.cursorId = findReply.Cursor.Id
			}
		} else {
			rdocs := int(op.replyDocs)
			if docNum == 0 {
				iter.docsToReceive += rdocs - 1
				docsToProcess := iter.docData.Len() + rdocs
				if iter.limit == 0 || int32(docsToProcess) < iter.limit {
					iter.docsBeforeMore = docsToProcess - int(iter.prefetch*float64(rdocs))
				} else {
					iter.docsBeforeMore = -1
				}
				iter.op.cursorId = op.cursorId
			}
			debugf("Iter %p received reply document %d/%d (cursor=%d)", iter, docNum+1, rdocs, op.cursorId)
			iter.docData.Push(docData)
		}
		iter.gotReply.Broadcast()
		iter.m.Unlock()
	}
}

type writeCmdResult struct {
	Ok        bool
	N         int
	NModified int `bson:"nModified"`
	Upserted  []struct {
		Index int
		Id    interface{} `_id`
	}
	ConcernError writeConcernError `bson:"writeConcernError"`
	Errors       []writeCmdError   `bson:"writeErrors"`
}

type writeConcernError struct {
	Code   int
	ErrMsg string
}

type writeCmdError struct {
	Index  int
	Code   int
	ErrMsg string
}

func (r *writeCmdResult) BulkErrorCases() []BulkErrorCase {
	ecases := make([]BulkErrorCase, len(r.Errors))
	for i, err := range r.Errors {
		ecases[i] = BulkErrorCase{err.Index, &QueryError{Code: err.Code, Message: err.ErrMsg}}
	}
	return ecases
}

// writeOp runs the given modifying operation, potentially followed up
// by a getLastError command in case the session is in safe mode.  The
// LastError result is made available in lerr, and if lerr.Err is set it
// will also be returned as err.
func (c *Collection) writeOp(op interface{}, ordered bool) (lerr *LastError, err error) {
	s := c.Database.Session
	socket, err := s.acquireSocket(c.Database.Name == "local")
	if err != nil {
		return nil, err
	}
	defer socket.Release()

	s.m.RLock()
	safeOp := s.safeOp
	bypassValidation := s.bypassValidation
	s.m.RUnlock()

	if socket.ServerInfo().MaxWireVersion >= 2 {
		// Servers with a more recent write protocol benefit from write commands.
		if op, ok := op.(*insertOp); ok && len(op.documents) > 1000 {
			var lerr LastError

			// Maximum batch size is 1000. Must split out in separate operations for compatibility.
			all := op.documents
			for i := 0; i < len(all); i += 1000 {
				l := i + 1000
				if l > len(all) {
					l = len(all)
				}
				op.documents = all[i:l]
				oplerr, err := c.writeOpCommand(socket, safeOp, op, ordered, bypassValidation)
				lerr.N += oplerr.N
				lerr.modified += oplerr.modified
				if err != nil {
					for ei := range lerr.ecases {
						oplerr.ecases[ei].Index += i
					}
					lerr.ecases = append(lerr.ecases, oplerr.ecases...)
					if op.flags&1 == 0 {
						return &lerr, err
					}
				}
			}
			if len(lerr.ecases) != 0 {
				return &lerr, lerr.ecases[0].Err
			}
			return &lerr, nil
		}
		return c.writeOpCommand(socket, safeOp, op, ordered, bypassValidation)
	} else if updateOps, ok := op.(bulkUpdateOp); ok {
		var lerr LastError
		for i, updateOp := range updateOps {
			oplerr, err := c.writeOpQuery(socket, safeOp, updateOp, ordered)
			lerr.N += oplerr.N
			lerr.modified += oplerr.modified
			if err != nil {
				lerr.ecases = append(lerr.ecases, BulkErrorCase{i, err})
				if ordered {
					break
				}
			}
		}
		if len(lerr.ecases) != 0 {
			return &lerr, lerr.ecases[0].Err
		}
		return &lerr, nil
	} else if deleteOps, ok := op.(bulkDeleteOp); ok {
		var lerr LastError
		for i, deleteOp := range deleteOps {
			oplerr, err := c.writeOpQuery(socket, safeOp, deleteOp, ordered)
			lerr.N += oplerr.N
			lerr.modified += oplerr.modified
			if err != nil {
				lerr.ecases = append(lerr.ecases, BulkErrorCase{i, err})
				if ordered {
					break
				}
			}
		}
		if len(lerr.ecases) != 0 {
			return &lerr, lerr.ecases[0].Err
		}
		return &lerr, nil
	}
	return c.writeOpQuery(socket, safeOp, op, ordered)
}

func (c *Collection) writeOpQuery(socket *mongoSocket, safeOp *queryOp, op interface{}, ordered bool) (lerr *LastError, err error) {
	if safeOp == nil {
		return nil, socket.Query(op)
	}

	var mutex sync.Mutex
	var replyData []byte
	var replyErr error
	mutex.Lock()
	query := *safeOp // Copy the data.
	query.collection = c.Database.Name + ".$cmd"
	query.replyFunc = func(err error, reply *replyOp, docNum int, docData []byte) {
		replyData = docData
		replyErr = err
		mutex.Unlock()
	}
	err = socket.Query(op, &query)
	if err != nil {
		return nil, err
	}
	mutex.Lock() // Wait.
	if replyErr != nil {
		return nil, replyErr // XXX TESTME
	}
	if hasErrMsg(replyData) {
		// Looks like getLastError itself failed.
		err = checkQueryError(query.collection, replyData)
		if err != nil {
			return nil, err
		}
	}
	result := &LastError{}
	bson.Unmarshal(replyData, &result)
	debugf("Result from writing query: %#v", result)
	if result.Err != "" {
		result.ecases = []BulkErrorCase{{Index: 0, Err: result}}
		if insert, ok := op.(*insertOp); ok && len(insert.documents) > 1 {
			result.ecases[0].Index = -1
		}
		return result, result
	}
	// With MongoDB <2.6 we don't know how many actually changed, so make it the same as matched.
	result.modified = result.N
	return result, nil
}

func (c *Collection) writeOpCommand(socket *mongoSocket, safeOp *queryOp, op interface{}, ordered, bypassValidation bool) (lerr *LastError, err error) {
	var writeConcern interface{}
	if safeOp == nil {
		writeConcern = bson.D{{"w", 0}}
	} else {
		writeConcern = safeOp.query.(*getLastError)
	}

	var cmd bson.D
	switch op := op.(type) {
	case *insertOp:
		// http://docs.mongodb.org/manual/reference/command/insert
		cmd = bson.D{
			{"insert", c.Name},
			{"documents", op.documents},
			{"writeConcern", writeConcern},
			{"ordered", op.flags&1 == 0},
		}
	case *updateOp:
		// http://docs.mongodb.org/manual/reference/command/update
		cmd = bson.D{
			{"update", c.Name},
			{"updates", []interface{}{op}},
			{"writeConcern", writeConcern},
			{"ordered", ordered},
		}
	case bulkUpdateOp:
		// http://docs.mongodb.org/manual/reference/command/update
		cmd = bson.D{
			{"update", c.Name},
			{"updates", op},
			{"writeConcern", writeConcern},
			{"ordered", ordered},
		}
	case *deleteOp:
		// http://docs.mongodb.org/manual/reference/command/delete
		cmd = bson.D{
			{"delete", c.Name},
			{"deletes", []interface{}{op}},
			{"writeConcern", writeConcern},
			{"ordered", ordered},
		}
	case bulkDeleteOp:
		// http://docs.mongodb.org/manual/reference/command/delete
		cmd = bson.D{
			{"delete", c.Name},
			{"deletes", op},
			{"writeConcern", writeConcern},
			{"ordered", ordered},
		}
	}
	if bypassValidation {
		cmd = append(cmd, bson.DocElem{"bypassDocumentValidation", true})
	}

	var result writeCmdResult
	err = c.Database.run(socket, cmd, &result)
	debugf("Write command result: %#v (err=%v)", result, err)
	ecases := result.BulkErrorCases()
	lerr = &LastError{
		UpdatedExisting: result.N > 0 && len(result.Upserted) == 0,
		N:               result.N,

		modified: result.NModified,
		ecases:   ecases,
	}
	if len(result.Upserted) > 0 {
		lerr.UpsertedId = result.Upserted[0].Id
	}
	if len(result.Errors) > 0 {
		e := result.Errors[0]
		lerr.Code = e.Code
		lerr.Err = e.ErrMsg
		err = lerr
	} else if result.ConcernError.Code != 0 {
		e := result.ConcernError
		lerr.Code = e.Code
		lerr.Err = e.ErrMsg
		err = lerr
	}

	if err == nil && safeOp == nil {
		return nil, nil
	}
	return lerr, err
}

func hasErrMsg(d []byte) bool {
	l := len(d)
	for i := 0; i+8 < l; i++ {
		if d[i] == '\x02' && d[i+1] == 'e' && d[i+2] == 'r' && d[i+3] == 'r' && d[i+4] == 'm' && d[i+5] == 's' && d[i+6] == 'g' && d[i+7] == '\x00' {
			return true
		}
	}
	return false
}
