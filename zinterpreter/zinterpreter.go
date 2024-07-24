package zinterpreter

// zanzibar restricted BNF grammar
/**
<Zschema> ::= <Zdef>*
<Zdef> ::= "definition" <Zname> "{" <Zbody> "}"
<Zname> ::= <identifier>
<Zbody> ::= <Zrelation>*
<Zrelation> ::= "relation" <Zname> ":" <Zname> ("|" <ZName)*
<Zname> ::= <identifier>
<identifier> ::= [a-zA-Z_][a-zA-Z0-9_]*

*/

import (
	"fmt"
	"strings"
	"unicode"
)

// Token represents the different tokens
type Token int

const (
	DefinitionToken Token = iota // "definition"
	RelationToken                // "relation"
	ColonToken                   // ":"
	OrToken                      // "|"
	LeftBraceToken               // "{"
	RightBraceToken              // "}"
	IdentifierToken              // [a-zA-Z_][a-zA-Z0-9_]*
	EOFToken                     // ''
	InvalidToken                 //
)

// Item représente un token avec sa valeur
type Item struct {
	Token Token
	Value string
}

// Lexer parses input text and generates tokens
type Lexer struct {
	input       string
	pos         int
	length      int
	currentItem *Item
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		length: len(input),
		currentItem: &Item{
			Token: InvalidToken,
			Value: "",
		},
	}
}

// We eat up the white spaces
func (l *Lexer) eatSpace() {
	for l.pos < l.length && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

// Lexer returns the next token to read
func (l *Lexer) NextToken() *Item {
	l.eatSpace()

	if l.pos >= l.length {
		l.currentItem.Token = EOFToken
		l.currentItem.Value = ""
		return l.currentItem
	}

	switch {
	case strings.HasPrefix(l.input[l.pos:], "definition"):
		l.currentItem.Token = DefinitionToken
		l.currentItem.Value = "definition"
		l.pos += len("definition")
	case strings.HasPrefix(l.input[l.pos:], "relation"):
		l.currentItem.Token = RelationToken
		l.currentItem.Value = "relation"
		l.pos += len("relation")
	case l.input[l.pos] == ':':
		l.currentItem.Token = ColonToken
		l.currentItem.Value = ":"
		l.pos++
	case l.input[l.pos] == '|':
		l.currentItem.Token = OrToken
		l.currentItem.Value = "|"
		l.pos++
	case l.input[l.pos] == '{':
		l.currentItem.Token = LeftBraceToken
		l.currentItem.Value = "{"
		l.pos++
	case l.input[l.pos] == '}':
		l.currentItem.Token = RightBraceToken
		l.currentItem.Value = "}"
		l.pos++
	default:
		if unicode.IsLetter(rune(l.input[l.pos])) {
			start := l.pos
			for l.pos < l.length && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
				l.pos++
			}
			l.currentItem.Token = IdentifierToken
			l.currentItem.Value = l.input[start:l.pos]
		} else {
			l.currentItem.Token = InvalidToken
			l.currentItem.Value = string(l.input[l.pos])
			l.pos++
		}
	}
	return l.currentItem
}

func (l *Lexer) readAndMatchToken(expected Token) error {
	if l.currentItem.Token == expected {
		return nil
	}
	return fmt.Errorf("expected token '%v', but got '%v'", expected, l.currentItem.Value)
}

// Syntaxic Analyser
type ZDef struct {
	Name      string
	Relations []ZRelation
}

type ZRelation struct {
	Name  string
	Types []string
}

// <Zschema> ::= <Zdef>*
func (l *Lexer) ReadZSchema() ([]ZDef, error) {
	var zdefs []ZDef

	for l.currentItem.Token != EOFToken {
		_zdef, _err := l.readZDef()
		if _err != nil {
			return zdefs, _err
		}
		zdefs = append(zdefs, _zdef)
		l.NextToken()

	}
	return zdefs, nil
}

// <Zdef> ::= "definition" <Zname> "{" <Zbody> "}"
func (l *Lexer) readZDef() (ZDef, error) {
	var zdef ZDef

	// read "definition"
	err := l.readAndMatchToken(DefinitionToken)
	if err != nil {
		return zdef, err
	}
	l.NextToken()

	// read <Zname>
	err = l.readAndMatchToken(IdentifierToken)
	if err != nil {
		return zdef, err
	}
	zdef.Name = l.currentItem.Value
	l.NextToken()

	// read '{'
	err = l.readAndMatchToken(LeftBraceToken)
	if err != nil {
		return zdef, err
	}
	l.NextToken()

	// read ZBody
	// ZBody is not a token
	// no need to call NextToken after

	zdef, err = l.readZBody(zdef)
	if err != nil {
		return zdef, err
	}

	// read '}'
	err = l.readAndMatchToken(RightBraceToken)
	if err != nil {
		return zdef, err
	}

	return zdef, nil
}

// <Zbody> ::= <Zrelation>*
// * means zero or more <Zrelation>
func (l *Lexer) readZBody(zdef ZDef) (ZDef, error) {
	// var zdef ZDef

	for l.currentItem.Token == RelationToken {

		relation, err := l.readZRelation()
		if err != nil {
			return zdef, err
		}
		zdef.Relations = append(zdef.Relations, relation)
	}

	return zdef, nil
}

// <Zrelation> ::= "relation" <Zname> ":" <Zname> ("|" <Zname>)*
func (l *Lexer) readZRelation() (ZRelation, error) {
	var zrelation ZRelation

	if l.currentItem.Value != "relation" {
		return zrelation, fmt.Errorf("expected 'relation', but got '%s'", l.currentItem.Value)
	}
	l.NextToken()

	err := l.readAndMatchToken(IdentifierToken)
	if err != nil {
		return zrelation, err
	}
	zrelation.Name = l.currentItem.Value
	l.NextToken()

	err = l.readAndMatchToken(ColonToken)
	if err != nil {
		return zrelation, err
	}
	l.NextToken()

	for l.currentItem.Token == IdentifierToken {
		zrelation.Types = append(zrelation.Types, l.currentItem.Value)
		l.NextToken()

		if l.currentItem.Token == OrToken {
			l.NextToken()
		} else {
			break
			// it's the last
		}
	}
	return zrelation, nil
}

// Archimate PlantUML generation code part II

type ArchimateRelation struct {
	subjectVar   string
	objectVar    string
	relationName string
}

func GenPlantUmlArchimate(zdefs []ZDef, pngfilename string) string {
	var out []string
	businessObjects := createZDefMap(zdefs)
	businessRelations := createArchimateRelations(zdefs, businessObjects)

	// fmt.Println(businessObjects)
	// fmt.Println(businessRelations)
	out = append(out, "@startuml "+pngfilename)
	out = append(out, "!include <archimate/Archimate>")

	// Générer une ligne pour chaque businessObject
	for name, varName := range businessObjects {
		line := fmt.Sprintf("Business_Object(%s,\"%s\")", varName, name)
		out = append(out, line)

	}

	// Générer une ligne pour chaque businessRelation
	for _, relation := range businessRelations {
		line := fmt.Sprintf("Rel_Access_w(%s,%s,\"%s\")", relation.subjectVar, relation.objectVar, relation.relationName)
		out = append(out, line)
	}

	out = append(out, "@enduml")
	return strings.Join(out, "\n")

}

func createArchimateRelations(zdefs []ZDef, bmap map[string]string) []ArchimateRelation {
	var relations []ArchimateRelation
	// Parcourir tous les ZDefs pour créer les relations
	for _, zdef := range zdefs {
		objectVar := bmap[zdef.Name]
		for _, zrel := range zdef.Relations {
			for _, relType := range zrel.Types {
				subjectVar := bmap[relType]
				relations = append(relations, ArchimateRelation{
					subjectVar:   subjectVar,
					objectVar:    objectVar,
					relationName: zrel.Name,
				})
			}
		}
	}
	return relations
}

func createZDefMap(zdefs []ZDef) map[string]string {
	zdefMap := make(map[string]string)
	for index, zdef := range zdefs {
		varname := fmt.Sprintf("b%d", index+1)
		if _, exists := zdefMap[zdef.Name]; exists {
			fmt.Printf("Please correct the schema file. Definition %s is duplicate.  \n", zdef.Name)
			continue
		}
		zdefMap[zdef.Name] = varname
	}
	return zdefMap
}
