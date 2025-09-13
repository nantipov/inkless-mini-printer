package ipp

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/OpenPrinting/goipp"

	"nantipov/printerservice/internal"
	"nantipov/printerservice/internal/jobprocessor"
)

var (
	ippFormatter = goipp.NewFormatter()

	staticAttibutes = goipp.Attributes{
		goipp.MakeAttribute("printer-name", goipp.TagName, goipp.String("Milcus")),
		goipp.MakeAttribute("printer-info", goipp.TagText, goipp.String("Inkless printer")),
		goipp.MakeAttribute("device-uri", goipp.TagURI, goipp.String("ipp://localhost:8080/ipp/print")),
		goipp.MakeAttribute("printer-is-accepting-jobs", goipp.TagBoolean, goipp.Boolean(true)),
		goipp.MakeAttribute("printer-location", goipp.TagText, goipp.String("Wohnzimmer")),
		goipp.MakeAttribute("printer-dns-sd-name", goipp.TagName, goipp.String("local")),
		goipp.MakeAttribute("printer-make-and-model", goipp.TagText, goipp.String("MilcusA1")),
		goipp.MakeAttribute("printer-more-info", goipp.TagURI, goipp.String("http://localhost:8080/info")),
		goipp.MakeAttribute("printer-uuid", goipp.TagText, goipp.String("8b05f7a4-6c79-413f-b68c-54700db846630")),
		goipp.MakeAttribute("uri-authentication-supported", goipp.TagKeyword, goipp.String("none")),

		goipp.MakeAttribute("document-format-supported", goipp.TagMimeType, goipp.String("application/pdf")),
		goipp.MakeAttribute("document-format-default", goipp.TagMimeType, goipp.String("application/pdf")),

		goipp.MakeAttribute("compression-supported", goipp.TagKeyword, goipp.String("none")),
		goipp.MakeAttribute("copies-supported", goipp.TagInteger, goipp.Integer(1)),

		goipp.MakeAttribute("job-password-encryption-supported", goipp.TagKeyword, goipp.String("none")),
		goipp.MakeAttribute("job-presets-supported", goipp.TagKeyword, goipp.String("none")),

		goipp.MakeAttribute("printer-state", goipp.TagEnum, goipp.Integer(3)),
		goipp.MakeAttribute("printer-state-reasons", goipp.TagKeyword, goipp.String("none")),

		goipp.MakeAttribute("operations-supported", goipp.TagEnum, goipp.Integer(goipp.OpPrintJob)),
	}

	staticAttibutesByNames = getAttibutesByNameMap(staticAttibutes)
)

// https://github.com/OpenPrinting/goipp/blob/master/doc.go

func HandleIpp(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if contentType != goipp.ContentType {
		http.Error(w, "unsupported content type", 400)
		return
	}

	ippReq := &goipp.Message{}
	err := ippReq.Decode(req.Body)

	ippFormatter.Reset()
	ippFormatter.FmtRequest(ippReq)
	log.Printf("Request \n %s", ippFormatter.String())

	if err != nil {
		http.Error(w, err.Error(), 400)
		// todo: log
		return
	}

	var ippResp *goipp.Message

	switch goipp.Op(ippReq.Code) {
	case goipp.OpGetPrinterAttributes:
		ippResp = getPrinterAttributes(ippReq)
	case goipp.OpPrintJob:
		ippResp = printJob(ippReq, req.Body)
	default:
		ippResp = goipp.NewResponse(goipp.DefaultVersion, goipp.StatusOk, ippReq.RequestID)
	}

	ippFormatter.Reset()
	ippFormatter.FmtResponse(ippResp)
	log.Printf("Response \n %s", ippFormatter.String())

	ippResp.Encode(w)
}

func printJob(request *goipp.Message, httpPayloadReader io.Reader) *goipp.Message {
	response := goipp.NewResponse(goipp.DefaultVersion, goipp.StatusOk, request.RequestID)

	response.Job.Add(goipp.MakeAttribute("job-id", goipp.TagInteger, goipp.Integer(1)))
	response.Job.Add(goipp.MakeAttribute("job-uri", goipp.TagURI, goipp.String("ipp://localhost:8080/ipp/job/1")))
	response.Job.Add(goipp.MakeAttribute("job-state", goipp.TagEnum, goipp.Integer(3)))
	response.Job.Add(goipp.MakeAttribute("job-state-reasons", goipp.TagKeyword, goipp.String("none")))
	response.Job.Add(goipp.MakeAttribute("job-state-message", goipp.TagText, goipp.String("Looking...")))

	f, err := os.Create("/tmp/f.pdf")
	internal.HandleError(err)

	written, err := io.Copy(f, httpPayloadReader)
	internal.HandleError(err)

	log.Printf("written %d", written)

	f.Close()

	job, _ := jobprocessor.CreateDraftJob(request.Job)
	// todo: handle error from create job
	job.AttachDocument("/tmp/f.pdf")

	return response
}

func getPrinterAttributes(request *goipp.Message) *goipp.Message {
	// https://datatracker.ietf.org/doc/html/rfc8011#section-5.4
	response := goipp.NewResponse(goipp.DefaultVersion, goipp.StatusOk, request.RequestID)

	attributesMap := getAttibutesByNameMap(request.Operation)

	requestedAttributes, requestedAttributesExists := attributesMap["requested-attributes"]
	if requestedAttributesExists {

		for _, requestedNameValue := range requestedAttributes.Values {
			requestedName := requestedNameValue.V.String()
			if requestedName == "all" {
				// add all
				//todo: attribute utils?
				for _, attribute := range staticAttibutes {
					response.Printer.Add(attribute)
				}
				break
			}
			attr, exists := staticAttibutesByNames[requestedName]
			if exists {
				response.Printer.Add(attr)
			}
		}
	} else {
		// add all
		for _, attribute := range staticAttibutes {
			response.Printer.Add(attribute)
		}
	}

	return response
}

func getAttibutesByNameMap(attributes goipp.Attributes) map[string]goipp.Attribute {
	attributesMap := make(map[string]goipp.Attribute)
	for _, attribute := range attributes {
		attributesMap[attribute.Name] = attribute
	}
	return attributesMap
}
