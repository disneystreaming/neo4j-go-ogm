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
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type SessionImpl struct {
	cypherExecuter *cypherExecuter
	saver          *saver
	loader         *loader
	deleter        *deleter
	queryer        *queryer
	transactioner  *transactioner
	store          store
	registry       *registry
	driver         neo4j.Driver
	eventer        *eventer
}

func LoadGeneric[MemberType Member](session *SessionImpl, object *MemberType, ID any, loadOptions *LoadOptions) error {
	return session.Load(object, ID, loadOptions)
}

func LoadAllGeneric[MemberType Member](session *SessionImpl, object *[]MemberType, ID any, loadOptions *LoadOptions) error {
	return session.LoadAll(object, ID, loadOptions)
}

func SaveGeneric[MemberType Member](session *SessionImpl, object *MemberType, saveOptions *SaveOptions) error {
	return session.Save(object, saveOptions)
}

func (s *SessionImpl) Load(object any, ID any, loadOptions *LoadOptions) error {

	_, err := s.loader.load(object, ID, loadOptions, false)
	return err
}

func (s *SessionImpl) LoadAll(objects any, IDs any, loadOptions *LoadOptions) error {
	return s.loader.loadAll(objects, IDs, loadOptions)
}

//TODO: Need to rework these functions back into instances and create generics for them
func (s *SessionImpl) Reload(objects ...any) error {
	return s.loader.reload(objects...)
}

func (s *SessionImpl) Save(objects any, saveOptions *SaveOptions) error {
	return s.saver.save(objects, saveOptions)
}

func (s *SessionImpl) Delete(object any) error {
	return s.deleter.delete(object)
}

func (s *SessionImpl) DeleteAll(objects any, deleteOptions *DeleteOptions) error {
	return s.deleter.deleteAll(objects, deleteOptions)
}

func (s *SessionImpl) PurgeDatabase() error {
	var err error
	if err = s.deleter.purgeDatabase(); err != nil {
		return err
	}
	return s.store.clear()
}

func (s *SessionImpl) Clear() error {
	return s.store.clear()
}

func (s *SessionImpl) BeginTransaction() (*Transaction, error) {
	return s.transactioner.beginTransaction(s)
}

func (s *SessionImpl) GetTransaction() *Transaction {
	return s.transactioner.transaction
}

//Precondition:
// * object is a pointer to a pointer of domain object: **<domainObject>
// * cypher returns one record with a column of domain object(s)
// * database entity type - node/relationhip - returned by cypher matches the domain object type - node/relationship
// * it is the user's resposibility to make sure  the database object returned by cypher are unloadable into domain object
//
//Post condition:
//Polulated domain objects
func (s *SessionImpl) QueryForObject(object any, cypher string, parameters map[string]any) error {
	return s.queryer.queryForObject(object, cypher, parameters)
}

//Precondition:
// * objects is a pointer to slice of pointers to domain objects: *[]*<domainObject>
// * cypher returns one or more record(s) with a column of domain object(s)
// * database entity type - node/relationhip - returned by cypher matches the domain object type -node/relationship
// * it is the user's resposibility to make sure that database objects returned by cypher are unloadable into the domain object
//
//Post condition:
//Polulated domain objects
func (s *SessionImpl) QueryForObjects(objects any, cypher string, parameters map[string]any) error {
	return s.queryer.queryForObjects(objects, cypher, parameters)
}

func (s *SessionImpl) Query(cypher string, parameters map[string]any, objects ...any) ([]map[string]any, error) {
	return s.queryer.query(cypher, parameters, objects...)
}

func (s *SessionImpl) CountEntitiesOfType(object any) (int64, error) {
	return s.queryer.countEntitiesOfType(object)
}

func (s *SessionImpl) Count(cypher string, parameters map[string]any) (int64, error) {
	return s.queryer.count(cypher, parameters)
}

func (s *SessionImpl) RegisterEventListener(eventListener EventListener) error {
	return s.eventer.registerEventListener(eventListener)
}
func (s *SessionImpl) DisposeEventListener(eventListener EventListener) error {
	return s.eventer.disposeEventListener(eventListener)
}
