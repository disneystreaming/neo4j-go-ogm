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
	"strconv"
	"strings"
)

type relationshipQueryBuilder struct {
	r               *relationship
	registry        *registry
	deltaProperties map[string]any
}

func (rqb relationshipQueryBuilder) getGraph() graph {
	return rqb.r
}

func newRelationshipCypherBuilder(r *relationship, registry *registry, stored graph) relationshipQueryBuilder {
	deltaProperties := r.getProperties()
	if stored != nil {
		deltaProperties = diffProperties(deltaProperties, stored.getProperties())
	}
	return relationshipQueryBuilder{
		r,
		registry,
		deltaProperties}
}

func (rqb relationshipQueryBuilder) getRemovedGraphs() (map[int64]graph, map[int64]graph) {
	return nil, nil
}

func (rqb relationshipQueryBuilder) isGraphDirty() bool {
	return rqb.r.getID() < 0 || len(rqb.deltaProperties) > 0
}

func (rqb relationshipQueryBuilder) getCreate() (string, string, map[string]any, map[string]graph) {
	var (
		r         = rqb.r
		startSign = r.nodes[startNode].getSignature()
		endSign   = r.nodes[endNode].getSignature()
		rSign     = r.getSignature()
	)
	create := `CREATE (` + startSign + `)-[` + rSign + `:` + r.getType() + `]->(` + endSign + `)
	`
	return "", create, nil, map[string]graph{startSign: r.nodes[startNode], endSign: r.nodes[endNode]}
}

func (rqb relationshipQueryBuilder) getMatch() (string, map[string]any, map[string]graph) {
	var (
		r         = rqb.r
		startSign = r.nodes[startNode].getSignature()
		endSign   = r.nodes[endNode].getSignature()
		rSign     = r.getSignature()
		match     = `MATCH (` + startSign + `)-[` + rSign + `:` + r.getType() + `]->(` + endSign + `)
		`
	)
	return match, nil, map[string]graph{startSign: r.nodes[startNode], endSign: r.nodes[endNode]}
}

func (rqb relationshipQueryBuilder) getSet() (string, map[string]any) {
	var (
		r          = rqb.r
		rSign      = r.getSignature()
		properties = map[string]any{}
		parameters = map[string]any{}
		propCQLRef = rSign + "Properties"
		set        string
	)
	for propertyName, propertyValue := range r.getProperties() {
		if !metaProperties[propertyName] {
			properties[propertyName] = propertyValue
		}
	}

	if len(properties) > 0 {
		set += `SET ` + rSign + ` += $` + propCQLRef + `
		`
		parameters[propCQLRef] = properties
	}

	return set, parameters
}

func (rqb relationshipQueryBuilder) getLoadAll(IDs any, lo *LoadOptions) (string, map[string]any) {

	var (
		depth                   = strconv.Itoa(lo.Depth)
		metadata, _             = rqb.registry.get(rqb.r.getValue().Type())
		customIDPropertyName, _ = metadata.getCustomID(*rqb.r.getValue())
		parameters              = map[string]any{}
	)

	if lo.Depth == infiniteDepth {
		depth = ""
	}
	matchOutString := fmt.Sprintf(`MATCH path = ()-[*0..%s]->()-[r:%s]->()-[*0..%s]->()`, depth, rqb.r.getLabel(), depth)
	matchInString := fmt.Sprintf(`MATCH path = ()<-[*0..%s]-()<-[r:%s]-()<-[*0..%s]-()`, depth, rqb.r.getLabel(), depth)
	unionString := `UNION`

	var filter string
	if IDs != nil {
		filter = `WHERE ID(r) IN $ids 
		`
		if customIDPropertyName != emptyString {
			filter = `WHERE r.` + customIDPropertyName + ` IN $ids 
			`
		}
		parameters["ids"] = IDs
	}

	end := `WITH r, path, range(0, length(path) - 1) as index
	WITH  r, path, index, [i in index | CASE WHEN nodes(path)[i] = startNode(relationships(path)[i]) THEN false ELSE true END] as isDirectionInverted
	RETURN path, ID(r), isDirectionInverted
	`
	match := fmt.Sprintf("%s", strings.Join([]string{matchOutString, filter, end, unionString, matchInString, filter, end}, " "))
	return match, parameters
}

func (rqb relationshipQueryBuilder) getDeleteAll() (string, map[string]any) {
	return `MATCH ()-[r:` + rqb.r.getType() + `]-()
	DELETE r
	RETURN ID(r)`, nil
}

func (rqb relationshipQueryBuilder) getDelete() (string, map[string]any, map[string]graph) {
	rSign := rqb.r.getSignature()
	delete, _, depedencies := rqb.getMatch()
	delete += `DELETE ` + rSign + ` RETURN ID(` + rSign + `)
	`
	return delete, nil, depedencies
}

func (rqb relationshipQueryBuilder) getCountEntitiesOfType() (string, map[string]any) {
	return `MATCH ()-[r:` + rqb.r.getType() + `]->() RETURN count(r) as count`, nil
}
