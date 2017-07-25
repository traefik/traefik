/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner_test

import (
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func ExampleNewClient() {
	ctx := context.Background()
	const myDB = "projects/my-project/instances/my-instance/database/my-db"
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	_ = client // TODO: Use client.
}

const myDB = "projects/my-project/instances/my-instance/database/my-db"

func ExampleNewClientWithConfig() {
	ctx := context.Background()
	const myDB = "projects/my-project/instances/my-instance/database/my-db"
	client, err := spanner.NewClientWithConfig(ctx, myDB, spanner.ClientConfig{
		NumChannels: 10,
	})
	if err != nil {
		// TODO: Handle error.
	}
	_ = client     // TODO: Use client.
	client.Close() // Close client when done.
}

func ExampleClient_Single() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().Query(ctx, spanner.NewStatement("SELECT FirstName FROM Singers"))
	_ = iter // TODO: iterate using Next or Do.
}

func ExampleClient_ReadOnlyTransaction() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	t := client.ReadOnlyTransaction()
	defer t.Close()
	// TODO: Read with t using Read, ReadRow, ReadUsingIndex, or Query.
}

func ExampleClient_ReadWriteTransaction() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	_, err = client.ReadWriteTransaction(ctx, func(txn *spanner.ReadWriteTransaction) error {
		var balance int64
		row, err := txn.ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"balance"})
		if err != nil {
			// This function will be called again if this is an
			// IsAborted error.
			return err
		}
		if err := row.Column(0, &balance); err != nil {
			return err
		}

		if balance <= 10 {
			return errors.New("insufficient funds in account")
		}
		balance -= 10
		m := spanner.Update("Accounts", []string{"user", "balance"}, []interface{}{"alice", balance})
		txn.BufferWrite([]*spanner.Mutation{m})

		// The buffered mutation will be committed.  If the commit
		// fails with an IsAborted error, this function will be called
		// again.
		return nil
	})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_Apply() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	m := spanner.Update("Users", []string{"name", "email"}, []interface{}{"alice", "a@example.com"})
	_, err = client.Apply(ctx, []*spanner.Mutation{m})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleInsert() {
	m := spanner.Insert("Users", []string{"name", "email"}, []interface{}{"alice", "a@example.com"})
	_ = m // TODO: use with Client.Apply or in a ReadWriteTransaction.
}

func ExampleInsertMap() {
	m := spanner.InsertMap("Users", map[string]interface{}{
		"name":  "alice",
		"email": "a@example.com",
	})
	_ = m // TODO: use with Client.Apply or in a ReadWriteTransaction.
}

func ExampleInsertStruct() {
	type User struct {
		Name, Email string
	}
	u := User{Name: "alice", Email: "a@example.com"}
	m, err := spanner.InsertStruct("Users", u)
	if err != nil {
		// TODO: Handle error.
	}
	_ = m // TODO: use with Client.Apply or in a ReadWriteTransaction.
}

func ExampleDelete() {
	m := spanner.Delete("Users", spanner.Key{"alice"})
	_ = m // TODO: use with Client.Apply or in a ReadWriteTransaction.
}

func ExampleDeleteKeyRange() {
	m := spanner.DeleteKeyRange("Users", spanner.KeyRange{
		Start: spanner.Key{"alice"},
		End:   spanner.Key{"bob"},
		Kind:  spanner.ClosedClosed,
	})
	_ = m // TODO: use with Client.Apply or in a ReadWriteTransaction.
}

func ExampleRowIterator_Next() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().Query(ctx, spanner.NewStatement("SELECT FirstName FROM Singers"))
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		var firstName string
		if err := row.Column(0, &firstName); err != nil {
			// TODO: Handle error.
		}
		fmt.Println(firstName)
	}
}

func ExampleRowIterator_Do() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().Query(ctx, spanner.NewStatement("SELECT FirstName FROM Singers"))
	err = iter.Do(func(r *spanner.Row) error {
		var firstName string
		if err := r.Column(0, &firstName); err != nil {
			return err
		}
		fmt.Println(firstName)
		return nil
	})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleRow_Size() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(row.Size()) // size is 2
}

func ExampleRow_ColumnName() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(row.ColumnName(1)) // prints "balance"
}

func ExampleRow_ColumnIndex() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	index, err := row.ColumnIndex("balance")
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(index)
}

func ExampleRow_ColumnNames() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(row.ColumnNames())
}

func ExampleRow_ColumnByName() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	var balance int64
	if err := row.ColumnByName("balance", &balance); err != nil {
		// TODO: Handle error.
	}
	fmt.Println(balance)
}

func ExampleRow_Columns() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}
	var name string
	var balance int64
	if err := row.Columns(&name, &balance); err != nil {
		// TODO: Handle error.
	}
	fmt.Println(name, balance)
}

func ExampleRow_ToStruct() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Accounts", spanner.Key{"alice"}, []string{"name", "balance"})
	if err != nil {
		// TODO: Handle error.
	}

	type Account struct {
		Name    string
		Balance int64
	}

	var acct Account
	if err := row.ToStruct(&acct); err != nil {
		// TODO: Handle error.
	}
	fmt.Println(acct)
}

func ExampleReadOnlyTransaction_Read() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().Read(ctx, "Users",
		spanner.Keys(spanner.Key{"alice"}, spanner.Key{"bob"}),
		[]string{"name", "email"})
	_ = iter // TODO: iterate using Next or Do.
}

func ExampleReadOnlyTransaction_ReadUsingIndex() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().ReadUsingIndex(ctx, "Users",
		"UsersByEmail",
		spanner.Keys(spanner.Key{"a@example.com"}, spanner.Key{"b@example.com"}),
		[]string{"name", "email"})
	_ = iter // TODO: iterate using Next or Do.
}

func ExampleReadOnlyTransaction_ReadRow() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	row, err := client.Single().ReadRow(ctx, "Users", spanner.Key{"alice"},
		[]string{"name", "email"})
	if err != nil {
		// TODO: Handle error.
	}
	_ = row // TODO: use row
}

func ExampleReadOnlyTransaction_Query() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	iter := client.Single().Query(ctx, spanner.NewStatement("SELECT FirstName FROM Singers"))
	_ = iter // TODO: iterate using Next or Do.
}

func ExampleNewStatement() {
	stmt := spanner.NewStatement("SELECT FirstName, LastName FROM SINGERS WHERE LastName >= @start")
	stmt.Params["start"] = "Dylan"
	// TODO: Use stmt in Query.
}

func ExampleNewStatement_structLiteral() {
	stmt := spanner.Statement{
		SQL:    "SELECT FirstName, LastName FROM SINGERS WHERE LastName >= @start",
		Params: map[string]interface{}{"start": "Dylan"},
	}
	_ = stmt // TODO: Use stmt in Query.
}

func ExampleReadOnlyTransaction_Timestamp() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	txn := client.Single()
	row, err := txn.ReadRow(ctx, "Users", spanner.Key{"alice"},
		[]string{"name", "email"})
	if err != nil {
		// TODO: Handle error.
	}
	readTimestamp, err := txn.Timestamp()
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println("read happened at", readTimestamp)
	_ = row // TODO: use row
}

func ExampleReadOnlyTransaction_WithTimestampBound() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, myDB)
	if err != nil {
		// TODO: Handle error.
	}
	txn := client.Single().WithTimestampBound(spanner.MaxStaleness(30 * time.Second))
	row, err := txn.ReadRow(ctx, "Users", spanner.Key{"alice"}, []string{"name", "email"})
	if err != nil {
		// TODO: Handle error.
	}
	_ = row // TODO: use row
	readTimestamp, err := txn.Timestamp()
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println("read happened at", readTimestamp)
}
