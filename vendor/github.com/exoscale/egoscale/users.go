package egoscale

// User represents a User
type User struct {
	APIKey    string `json:"apikey,omitempty" doc:"the api key of the user"`
	Account   string `json:"account,omitempty" doc:"the account name of the user"`
	AccountID *UUID  `json:"accountid,omitempty" doc:"the account ID of the user"`
	Created   string `json:"created,omitempty" doc:"the date and time the user account was created"`
	Email     string `json:"email,omitempty" doc:"the user email address"`
	FirstName string `json:"firstname,omitempty" doc:"the user firstname"`
	ID        *UUID  `json:"id,omitempty" doc:"the user ID"`
	IsDefault bool   `json:"isdefault,omitempty" doc:"true if user is default, false otherwise"`
	LastName  string `json:"lastname,omitempty" doc:"the user lastname"`
	RoleID    *UUID  `json:"roleid,omitempty" doc:"the ID of the role"`
	RoleName  string `json:"rolename,omitempty" doc:"the name of the role"`
	RoleType  string `json:"roletype,omitempty" doc:"the type of the role"`
	SecretKey string `json:"secretkey,omitempty" doc:"the secret key of the user"`
	State     string `json:"state,omitempty" doc:"the user state"`
	Timezone  string `json:"timezone,omitempty" doc:"the timezone user was created in"`
	UserName  string `json:"username,omitempty" doc:"the user name"`
}

// ListRequest builds the ListUsers request
func (user User) ListRequest() (ListCommand, error) {
	req := &ListUsers{
		ID:       user.ID,
		UserName: user.UserName,
	}

	return req, nil
}

// RegisterUserKeys registers a new set of key of the given user
//
// NB: only the APIKey and SecretKey will be filled
type RegisterUserKeys struct {
	ID *UUID `json:"id" doc:"User id"`
	_  bool  `name:"registerUserKeys" description:"This command allows a user to register for the developer API, returning a secret key and an API key. This request is made through the integration API port, so it is a privileged command and must be made on behalf of a user. It is up to the implementer just how the username and password are entered, and then how that translates to an integration API request. Both secret key and API key should be returned to the user"`
}

// Response returns the struct to unmarshal
func (RegisterUserKeys) Response() interface{} {
	return new(User)
}

//go:generate go run generate/main.go -interface=Listable ListUsers

// ListUsers represents the search for Users
type ListUsers struct {
	ID       *UUID  `json:"id,omitempty" doc:"List user by ID."`
	Keyword  string `json:"keyword,omitempty" doc:"List by keyword"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
	State    string `json:"state,omitempty" doc:"List users by state of the user account."`
	UserName string `json:"username,omitempty" doc:"List user by the username"`
	_        bool   `name:"listUsers" description:"Lists user accounts"`
}

// ListUsersResponse represents a list of users
type ListUsersResponse struct {
	Count int    `json:"count"`
	User  []User `json:"user"`
}
