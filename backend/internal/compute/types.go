// Package compute 是数据处理层（Python）的 HTTP 客户端。
package compute

import "encoding/json"

// Cell 是晶胞参数：三条棱长（Å）与三个夹角（°）。
type Cell struct {
	EdgeAAngstrom    float64 `json:"edge_a_angstrom"`
	EdgeBAngstrom    float64 `json:"edge_b_angstrom"`
	EdgeCAngstrom    float64 `json:"edge_c_angstrom"`
	AngleAlphaDegree float64 `json:"angle_alpha_degree"`
	AngleBetaDegree  float64 `json:"angle_beta_degree"`
	AngleGammaDegree float64 `json:"angle_gamma_degree"`
}

// Candidate 是一个候选结构；rule 与 caveats 保留原始 JSON，落库与回传时不用重新编码。
type Candidate struct {
	Prototype          string          `json:"prototype"`
	CoordinationNumber int             `json:"coordination_number"`
	Rule               json.RawMessage `json:"rule"`
	Caveats            json.RawMessage `json:"caveats"`
	Cell               Cell            `json:"cell"`
	Atoms              json.RawMessage `json:"atoms"`
}

// Prediction 是数据处理层对一次预测的完整回答。
type Prediction struct {
	Formula    string      `json:"formula"`
	Family     string      `json:"family"`
	Candidates []Candidate `json:"candidates"`
}
