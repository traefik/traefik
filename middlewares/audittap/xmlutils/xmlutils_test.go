package xmlutils

import (
	"encoding/xml"
	"strings"

	"github.com/beevik/etree"
	"github.com/containous/traefik/middlewares/audittap/types"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXmlToMap(t *testing.T) {

	x := `
    <ElemA>
      <IAmEmpty />
      <ATextElem>AndTheText  </ATextElem>
      <ElemAChildren>
        <Child1 AnAttr="anattr" OtherAttr="otherattr">IAmChild1</Child1>
        <Child2>
          <EmptyGrandChild/>
          <GrandChild1>  IAmGrandChild1</GrandChild1>
          <GrandChild2>IAmGrandChild2</GrandChild2>
        </Child2>
      </ElemAChildren>
    </ElemA>
  `

	doc := etree.NewDocument()
	if err := doc.ReadFromString(x); err != nil {
		t.Fatal(err)
	}

	els := []*etree.Element{doc.Root()}
	m := XMLToDataMap(els, []string{})

	assert.IsType(t, types.DataMap{}, m.Get("ElemA"))
	l1 := m.GetDataMap("ElemA")
	assert.NotEmpty(t, l1)
	assert.Empty(t, l1.GetString("IAmEmpty"))
	assert.Equal(t, "AndTheText", l1.GetString("ATextElem"))

	l2 := l1.GetDataMap("ElemAChildren")
	assert.NotEmpty(t, l2)

	l3 := l2.GetDataMap("Child2")
	assert.Empty(t, l3.GetString("EmptyGrandChild"))
	assert.Equal(t, "IAmGrandChild1", l3.GetString("GrandChild1"))
	assert.Equal(t, "IAmGrandChild2", l3.GetString("GrandChild2"))
}

func TestXmlToMapExclusions(t *testing.T) {

	x := `
    <ElemA>
      <ATextElem>AndTheText</ATextElem>
			<OtherTextElem>OtherText</OtherTextElem>
      <ElemAChildren>
        <Child1 AnAttr="anattr" OtherAttr="otherattr">IAmChild1</Child1>
      </ElemAChildren>
    </ElemA>
  `

	doc := etree.NewDocument()
	if err := doc.ReadFromString(x); err != nil {
		t.Fatal(err)
	}

	els := []*etree.Element{doc.Root()}
	m := XMLToDataMap(els, []string{"ATextElem", "ElemAChildren"})

	l1 := m.GetDataMap("ElemA")
	assert.NotEmpty(t, l1)
	assert.Empty(t, l1.GetString("ATextElem"))
	assert.Empty(t, l1.Get("ElemAChildren"))
	assert.Equal(t, "OtherText", l1.GetString("OtherTextElem"))
}

func TestXmlToMapIgnoresNamespacing(t *testing.T) {

	x := `
    <NS1:ElemA>
      <NS2:ATextElem>AndTheText</NS2:ATextElem>
			<NS2:OtherTextElem>OtherText</NS2OtherTextElem>
      <NS2:ElemAChildren>
        <Child1 AnAttr="anattr" OtherAttr="otherattr">IAmChild1</Child1>
      </NS2:ElemAChildren>
    </NS1:ElemA>
  `

	doc := etree.NewDocument()
	if err := doc.ReadFromString(x); err != nil {
		t.Fatal(err)
	}

	els := []*etree.Element{doc.Root()}
	m := XMLToDataMap(els, []string{})

	l1 := m.GetDataMap("ElemA")
	assert.NotEmpty(t, l1)
	assert.Equal(t, "AndTheText", l1.GetString("ATextElem"))
	assert.Equal(t, "OtherText", l1.GetString("OtherTextElem"))
	assert.NotEmpty(t, l1.Get("ElemAChildren"))
	l2 := l1.GetDataMap("ElemAChildren")
	assert.Equal(t, types.DataMap{"AnAttr": "anattr", "OtherAttr": "otherattr", "Child1": "IAmChild1"}, l2.Get("Child1"))
}

func TestAttributeMapping(t *testing.T) {
	x := `
	<ElemA>
		<ElemB xmlns="omitme" eb1="elb_123">
			<ElemB1>BEE1</ElemB1>
		</ElemB>
		<ElemC ec1="elc_123" ec2="elc_456" ec3="">ABCDEF</ElemC>
	</ElemA>
	`

	doc := etree.NewDocument()
	if err := doc.ReadFromString(x); err != nil {
		t.Fatal(err)
	}

	els := []*etree.Element{doc.Root()}
	m := XMLToDataMap(els, []string{})

	l1 := m.GetDataMap("ElemA")
	assert.Equal(t, types.DataMap{"ElemB1": "BEE1", "eb1": "elb_123"}, l1.Get("ElemB"))
	assert.Equal(t, types.DataMap{"ElemC": "ABCDEF", "ec1": "elc_123", "ec2": "elc_456"}, l1.Get("ElemC"))
}

func TestElementInnerToDocument(t *testing.T) {

	x := `
	<ElemA>
		<ATextElem>AndTheText</ATextElem>
		<OtherTextElem>OtherText</OtherTextElem>
		<ElemAChildren>
			<Child1 AnAttr="anattr" OtherAttr="otherattr">IAmChild1</Child1>
		</ElemAChildren>
	</ElemA>
	`

	dec := xml.NewDecoder(strings.NewReader(x))
	root, err := ScrollToNextElement(dec)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ElementInnerToDocument(root, dec)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "ElemA", doc.Root().Tag)
	children := doc.Root().ChildElements()
	assert.Equal(t, "ATextElem", children[0].Tag)
	assert.Equal(t, "OtherTextElem", children[1].Tag)
	assert.Equal(t, "ElemAChildren", children[2].Tag)
	assert.Equal(t, "Child1", children[2].ChildElements()[0].Tag)
}

func TestScrollToNextElement(t *testing.T) {
	x := `
	<ElemA>
		<ElemB atb="b1">BBB</ElemB>
		<!-- a comment -->
		<ElemC atc="c2">CCC</ElemC>
	</ElemA>
	`
	dec := xml.NewDecoder(strings.NewReader(x))
	el1, _ := ScrollToNextElement(dec)
	el2, _ := ScrollToNextElement(dec)
	el3, _ := ScrollToNextElement(dec)
	assert.Equal(t, "ElemA", el1.Name.Local)
	assert.Equal(t, "ElemB", el2.Name.Local)
	assert.Equal(t, "ElemC", el3.Name.Local)
}
