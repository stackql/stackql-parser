/*
Copyright 2019 The Vitess Authors.

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

package sqlparser

import (
	"strings"

	"github.com/stackql/stackql-parser/go/vt/proto/vtrpc"
	"github.com/stackql/stackql-parser/go/vt/vterrors"
)

type setNormalizer struct {
	err error
}

func (n *setNormalizer) rewriteSetComingUp(cursor *Cursor) bool {
	set, ok := cursor.node.(*Set)
	if ok {
		for i, expr := range set.Exprs {
			exp, err := n.normalizeSetExpr(expr)
			if err != nil {
				n.err = err
				return false
			}
			set.Exprs[i] = exp
		}
	}
	return true
}

func (n *setNormalizer) normalizeSetExpr(in *SetExpr) (*SetExpr, error) {
	switch in.Name.at { // using switch so we can use break
	case DoubleAt:
		if in.Scope != "" {
			return nil, vterrors.Errorf(vtrpc.Code_INVALID_ARGUMENT, "cannot use scope and @@")
		}
		switch {
		case strings.HasPrefix(in.Name.Lowered(), "session."):
			in.Name = NewColIdent(in.Name.Lowered()[8:])
			in.Scope = SessionStr
		case strings.HasPrefix(in.Name.Lowered(), "global."):
			in.Name = NewColIdent(in.Name.Lowered()[7:])
			in.Scope = GlobalStr
		case strings.HasPrefix(in.Name.Lowered(), "vitess_metadata."):
			in.Name = NewColIdent(in.Name.Lowered()[16:])
			in.Scope = VitessMetadataStr
		default:
			in.Name.at = NoAt
			in.Scope = SessionStr
		}
		return in, nil
	case SingleAt:
		if in.Scope != "" {
			return nil, vterrors.Errorf(vtrpc.Code_INVALID_ARGUMENT, "cannot mix scope and user defined variables")
		}
		return in, nil
	case NoAt:
		switch in.Scope {
		case "":
			in.Scope = SessionStr
		case "local":
			in.Scope = SessionStr
		}
		return in, nil
	}
	panic("this should never happen")
}
