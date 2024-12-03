package main

func (*TemplateDataTree) nGetParentNode(n int64, t TemplateDataTree) any {
	if n > 0 {
       t.nGetParentNode(n-1,*t.Parent) 
	} 

	return t
}
