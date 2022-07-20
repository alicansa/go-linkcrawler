package crawler

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type XpathPolicyExecutorTest struct {
	policyExecutor *XPathPolicyExecutor
}

func TestXpathPolicyExecutor(t *testing.T) {
	xpe := NewPolicyExecutor("//a[@href[not(contains(.,'http')) and not(contains(.,'mailto:')) and not(contains(.,'tel:'))]]")
	xpet := XpathPolicyExecutorTest{
		policyExecutor: xpe,
	}

	t.Run("Test find nodes with policy and returns href values", xpet.testFindNodesWithPolicyAndReturnsHrefValues)
}

func (xpe *XpathPolicyExecutorTest) testFindNodesWithPolicyAndReturnsHrefValues(t *testing.T) {
	htmlContent := `<html>
		<div>
			<a href='test'>test</a>
			<div>
				<a href='test2'>some link</a>
			</div>
		</div>
	</html>`
	reader := strings.NewReader(htmlContent)
	readerCloser := io.NopCloser(reader)
	result, err := xpe.policyExecutor.Execute(readerCloser)

	assert.Nil(t, err)
	assert.Len(t, result, 2)

	expectedList := []string{"test", "test2"}
	assert.Equal(t, expectedList, result)
}
