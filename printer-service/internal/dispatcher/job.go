package dispatcher

type Job struct {
	Id    string
	IppId int
}

func (j Job) AttachDocument(filename string) {
	//todo: content-type?
}

func (j Job) MarkAsAssembled() {

}
