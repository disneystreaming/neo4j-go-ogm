// MIT License
//
// Copyright (c) 2020 codingfinest
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gogm

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type transactionExecuter func(work neo4j.TransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)

type cypherExecuter struct {
	driver      neo4j.Driver
	accessMode  neo4j.AccessMode
	transaction *Transaction
}

//Creates a new instance of a `cypherExecuter` from provided neo4j configuration parameters, using an optional `transaction`
func newCypherExecuter(driver neo4j.Driver, accessMode neo4j.AccessMode, t *Transaction) *cypherExecuter {
	return &cypherExecuter{driver, accessMode, nil}
}

//Executes a given cql statement using the provided params within the context of the provided `transactionExecuter`
func (c *cypherExecuter) execTransaction(te transactionExecuter, cql string, params map[string]any) ([]*neo4j.Record, error) {

	if records, err := te(func(tx neo4j.Transaction) (any, error) {

		if result, err := tx.Run(cql, params); err != nil {
			return nil, err
		} else {
			return result.Collect()
		}
	}); err != nil {
		return nil, err
	} else if resultAsRecords, isRecordSlice := records.([]*neo4j.Record); isRecordSlice {
		return resultAsRecords, nil
	} else {
		return nil, fmt.Errorf("records returned by query, but not in expected form")
	}
}

//Executes a given cql statements using the provided params within the context of the c's own state.
func (c *cypherExecuter) exec(cql string, params map[string]any) ([]*neo4j.Record, error) {
	var (
		result  neo4j.Result
		session neo4j.Session
		err     error
	)
	if c.transaction != nil {
		if result, err = c.transaction.run(cql, params); err != nil {
			return nil, err
		}
		return result.Collect()
	}

	session = c.driver.NewSession(neo4j.SessionConfig{
		AccessMode: c.accessMode,
	})

	transactionMode := session.ReadTransaction
	if c.accessMode == neo4j.AccessModeWrite {
		transactionMode = session.WriteTransaction
	}
	if result, resErr := c.execTransaction(transactionMode, cql, params); resErr != nil {
		session.Close()
		return nil, resErr
	} else {
		session.Close()
		return result, nil
	}
}

//Setter function for `transaction`.
func (c *cypherExecuter) setTransaction(transaction *Transaction) {
	c.transaction = transaction
}
