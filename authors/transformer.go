package authors

import (
	"encoding/base64"
	"encoding/xml"

	"github.com/pborman/uuid"
)

// AuthorTransformer struct
type AuthorTransformer struct {
}

// UnMarshallTaxonomy - unmarshal the XML of a taxonomy
func (*AuthorTransformer) UnMarshallTaxonomy(contents []byte) ([]interface{}, error) {
	t := taxonomy{}
	err := xml.Unmarshal(contents, &t)
	if err != nil {
		return nil, err
	}
	interfaces := make([]interface{}, len(t.Terms))
	for i, d := range t.Terms {
		interfaces[i] = d
	}
	return interfaces, nil
}

// UnMarshallTerm - unmarshal the XML of a TME term
func (*AuthorTransformer) UnMarshallTerm(content []byte) (interface{}, error) {
	dummyTerm := term{}
	err := xml.Unmarshal(content, &dummyTerm)
	if err != nil {
		return term{}, err
	}
	return dummyTerm, nil
}

func transformAuthor(tmeTerm term, taxonomyName string) author {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)
	authorUUID := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
	aliasList := buildAliasList(tmeTerm.Aliases, tmeTerm.CanonicalName)
	return author{
		UUID:      authorUUID,
		Name:      tmeTerm.CanonicalName,
		PrefLabel: tmeTerm.CanonicalName,
		AlternativeIdentifiers: alternativeIdentifiers{
			TME:   []string{tmeIdentifier},
			UUIDs: []string{authorUUID},
		},
		Type:    "Person",
		Aliases: aliasList,
	}
}

func buildTmeIdentifier(rawID string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawID))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(tmeTermTaxonomyName))
	return id + "-" + taxonomyName
}

func removeDuplicates(slice []string) []string {
	newSlice := []string{}
	seen := make(map[string]bool)
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			newSlice = append(newSlice, v)
			seen[v] = true
		}
	}
	return newSlice
}

func buildAliasList(aList aliases, canonicalName string) []string {
	aliasList := make([]string, len(aList.Alias))
	for k, v := range aList.Alias {
		aliasList[k] = v.Name
	}
	aliasList = append(aliasList, canonicalName)
	aliasList = removeDuplicates(aliasList)
	return aliasList
}
