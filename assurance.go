package lungo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// AssuranceSchemaVersion is the version of the assurance document this package reads.
const AssuranceSchemaVersion = 1

// Assurance is a program's assurance document (`assurance.json`): what its Lean code claims and
// proves of each export, the code each export's proofs do not cover (its trust), and the
// assumptions about the host its claims are conditional on. Every generated package carries
// it; `Assurance()` of a generated package returns it.
type Assurance struct {
	SchemaVersion  int                      `json:"schema_version"`
	Program        string                   `json:"program"`
	Provenance     AssuranceProvenance      `json:"provenance"`
	Library        *AssuranceLibrary        `json:"library"`
	Specifications []AssuranceSpecification `json:"specifications"`
	Facilities     []AssuranceFacility      `json:"facilities"`
	Assumptions    []AssuranceAssumption    `json:"assumptions"`
	Claims         []AssuranceClaim         `json:"claims"`
	Roles          []AssuranceRole          `json:"roles"`
	Exports        []AssuranceExport        `json:"exports"`
}

// AssuranceProvenance is what produced the document.
type AssuranceProvenance struct {
	LeanVersion  string `json:"lean_version"`
	LeanGithash  string `json:"lean_githash"`
	LungoVersion string `json:"lungo_version"`
	BirVersion   int    `json:"bir_version"`
	RuntimeABI   int    `json:"runtime_abi"`
}

// AssuranceLibrary is lungo's Lean library, as the program used it.
type AssuranceLibrary struct {
	Package       string `json:"package"`
	SchemaVersion int    `json:"schema_version"`
}

// AssurancePosition is a one-based line and column.
type AssurancePosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// AssuranceSource is where a declaration is, relative to its Lake package.
type AssuranceSource struct {
	Package string             `json:"package"`
	File    string             `json:"file"`
	Start   *AssurancePosition `json:"start"`
	End     *AssurancePosition `json:"end"`
}

// AssuranceSpecification is a specification the program's claims cite.
type AssuranceSpecification struct {
	Name        string           `json:"name"`
	Kind        string           `json:"kind"`
	Statement   string           `json:"statement"`
	Package     *string          `json:"package"`
	Fingerprint string           `json:"fingerprint"`
	Source      *AssuranceSource `json:"source"`
}

// AssuranceOperation is an operation of a facility.
type AssuranceOperation struct {
	Name        string  `json:"name"`
	Symbol      *string `json:"symbol"`
	Fingerprint *string `json:"fingerprint"`
}

// AssuranceFacility is a facility the host provides.
type AssuranceFacility struct {
	Name        string               `json:"name"`
	ID          string               `json:"id"`
	Form        string               `json:"form"`
	OpType      *string              `json:"op_type"`
	Operations  []AssuranceOperation `json:"operations"`
	Assumptions []string             `json:"assumptions"`
	Package     *string              `json:"package"`
	Fingerprint string               `json:"fingerprint"`
	Source      *AssuranceSource     `json:"source"`
}

// AssuranceAssumption is a proposition assumed, never proved, of the host's implementation of a
// facility.
type AssuranceAssumption struct {
	Name        string           `json:"name"`
	Facility    string           `json:"facility"`
	Statement   string           `json:"statement"`
	Package     *string          `json:"package"`
	Fingerprint string           `json:"fingerprint"`
	Source      *AssuranceSource `json:"source"`
}

// AssuranceEvidenceTrust is what the evidence of a claim depends on.
type AssuranceEvidenceTrust struct {
	Axioms         []string `json:"axioms"`
	DependsOnSorry bool     `json:"depends_on_sorry"`
}

// AssuranceClaim is a theorem (its name is the claim's) proving that its subjects stand in a
// relation to its specifications. Status is "proved" or "incomplete" (resting on sorry).
type AssuranceClaim struct {
	Name           string                 `json:"name"`
	Relation       string                 `json:"relation"`
	Subjects       []string               `json:"subjects"`
	Specifications []string               `json:"specifications"`
	Statement      string                 `json:"statement"`
	Status         string                 `json:"status"`
	EvidenceTrust  AssuranceEvidenceTrust `json:"evidence_trust"`
	Assumptions    []string               `json:"assumptions"`
	Package        *string                `json:"package"`
	Fingerprint    string                 `json:"fingerprint"`
	Source         *AssuranceSource       `json:"source"`
}

// AssuranceRole is what a declaration is for.
type AssuranceRole struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Exported bool   `json:"exported"`
}

// AssuranceTrust is the code an export's proofs do not cover.
type AssuranceTrust struct {
	Axioms              []string `json:"axioms"`
	DependsOnSorry      bool     `json:"depends_on_sorry"`
	UnsafeDependencies  []string `json:"unsafe_dependencies"`
	PartialDependencies []string `json:"partial_dependencies"`
	ExternDependencies  []string `json:"extern_dependencies"`
}

// AssuranceExport is what one export is, does and depends on.
type AssuranceExport struct {
	Name        string           `json:"name"`
	Module      string           `json:"module"`
	Async       bool             `json:"async"`
	Trust       AssuranceTrust   `json:"trust"`
	Claims      []string         `json:"claims"`
	Assumptions []string         `json:"assumptions"`
	Facilities  []string         `json:"facilities"`
	Roles       []string         `json:"roles"`
	Source      *AssuranceSource `json:"source"`
}

// ParseAssurance reads an assurance document, refusing one of another schema version or with
// fields this package does not know.
func ParseAssurance(data []byte) (*Assurance, error) {
	var version struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &version); err != nil {
		return nil, err
	}
	if version.SchemaVersion == nil {
		return nil, fmt.Errorf("lungo: an assurance document without a schema_version")
	}
	if *version.SchemaVersion != AssuranceSchemaVersion {
		return nil, fmt.Errorf("lungo: assurance schema version %d; this package reads version %d", *version.SchemaVersion, AssuranceSchemaVersion)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var a Assurance
	if err := d.Decode(&a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Claim is the claim whose evidence is name.
func (a *Assurance) Claim(name string) *AssuranceClaim {
	for i := range a.Claims {
		if a.Claims[i].Name == name {
			return &a.Claims[i]
		}
	}
	return nil
}

// Export is the summary of the export name.
func (a *Assurance) Export(name string) *AssuranceExport {
	for i := range a.Exports {
		if a.Exports[i].Name == name {
			return &a.Exports[i]
		}
	}
	return nil
}
