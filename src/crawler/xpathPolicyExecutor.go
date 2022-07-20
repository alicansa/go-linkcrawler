package crawler

import (
	"io"

	"github.com/antchfx/htmlquery"
)

type XPathPolicyExecutor struct {
	Policy string
}

func (pe *XPathPolicyExecutor) Execute(rc io.ReadCloser) ([]string, error) {

	var output []string
	defer rc.Close()
	doc, err := htmlquery.Parse(rc)

	if err != nil {
		return output, err
	}

	nodes := htmlquery.Find(doc, pe.Policy)
	for _, node := range nodes {
		href := htmlquery.SelectAttr(node, "href")
		output = append(output, href)
	}
	return output, nil
}

func NewPolicyExecutor(policy string) *XPathPolicyExecutor {
	return &XPathPolicyExecutor{
		Policy: policy,
	}
}
