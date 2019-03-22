package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/resty.v1"
	"html/template"
	"net/url"

	//"math"
)
func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}


// Handler responds to http requests.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {



if (request.HTTPMethod=="GET") {

	assetid := request.QueryStringParameters["assetid"]

	resp, err := resty.R().Get("http://ec2-3-120-235-96.eu-central-1.compute.amazonaws.com:8000/hcms/v1.3/entity/product/" + assetid)

	// explore response object
	fmt.Printf("\nError: %v", err)
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())

	hedincar := Car{}
	json.Unmarshal([]byte(resp.String()), &hedincar)

	t := template.New("fieldname example")
	t, _ = t.Parse(`<html>

<!-- ******************************** HTML ****************************** -->

	<head>
		
	</head>
	<body>
	</form>
	</body>
</html>`)




	buf := new(bytes.Buffer)
	err = t.Execute(buf, hedincar)
	if err != nil {
		fmt.Println(err)
	}


	return events.APIGatewayProxyResponse{
		Body:       string(buf.Bytes()),
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "text/html; charset=utf-8",
			"Cache-Control":               "public, max-age=0, s-maxage=0, no-cache",
		},
	}, nil

} else {


	newcar:=&Car{}
m,err:=url.ParseQuery(request.Body)
if (err!=nil){
	fmt.Printf("\nError: %v", err)
}
fmt.Println(request.Body)
newcar.Regno=m["regno"][0]
newcar.Modeltext=m["modeltext"][0]

fmt.Println(newcar)

b,err:=json.Marshal(newcar)
if (err!=nil){
	fmt.Printf("\nError: %v", err)
}

fmt.Println(string(b))
fmt.Println("****START POST****")
resp, err := resty.R().
SetHeader("Content-Type", "application/json").
SetBody(string(b)).
Put("http://ec2-3-120-235-96.eu-central-1.compute.amazonaws.com:8000/hcms/v1.3/entity/product/119755")

if (err!=nil){
	fmt.Printf("\nError: %v", err)
}

fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())


return events.APIGatewayProxyResponse{
	Body:       string("OK"),
	StatusCode: 200,
	Headers: map[string]string{
		"Access-Control-Allow-Origin": "*",
		"Content-Type":                "text/html; charset=utf-8",
		"Cache-Control":               "public, max-age=0, s-maxage=0, no-cache",
	},
}, nil

}
}

func main() {
	lambda.Start(Handler)
}
