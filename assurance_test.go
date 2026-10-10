package lungo

import (
	"strings"
	"testing"
)

const assuranceDocument = `{"schema_version": 1, "program": "p",
  "provenance": {"lean_version": "4.34.1", "lean_githash": "x", "lungo_version": "1", "bir_version": 3, "runtime_abi": 2},
  "library": null,
  "specifications": [{"name": "P.s", "kind": "lungo.model", "statement": "Nat", "definition": null,
    "package": null, "fingerprint": "f", "source": null}],
  "facilities": [], "assumptions": [], "claims": [], "roles": [], "exports": []}`

// TEST0348: a document missing a field is refused, even a nullable one
func Test0348_ADocumentMissingAFieldIsRefusedEvenANullableOne(t *testing.T) {
	a, err := ParseAssurance([]byte(assuranceDocument))
	if err != nil || a.Specifications[0].Definition != nil {
		t.Fatalf("the document is refused: %v", err)
	}
	refused := func(text, want string) {
		t.Helper()
		_, err := ParseAssurance([]byte(text))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: %v", want, err)
		}
	}
	refused(strings.Replace(assuranceDocument, `"definition": null,`, "", 1), "document.specifications[0] lacks the field definition")
	refused(strings.Replace(assuranceDocument, `"library": null,`, "", 1), "document lacks the field library")
	refused(strings.Replace(assuranceDocument, `"roles": [],`, `"roles": [], "extra": 1,`, 1), "extra")
}
