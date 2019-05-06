package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	_ "golang.org/x/text/encoding/charmap"
	"gopkg.in/resty.v1"
	"log"
	_ "net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Brand struct {
	Name string `json:"name"`
}

type Car struct {
	Chassino           string       `json:"chassino"`
	Regno              string       `json:"regno"`
	Brand              string       `json:"brand"`
	ModelManual        string       `json:"modelManual"`
	Modeltext          string       `json:"modeltext"`
	RegistrationDate   int          `json:"registrationDate"`
	Yearmodel          string       `json:"yearmodel"`
	Mileage            int          `json:"mileage"`
	SalesPriceIncVAT   float64      `json:"salesPriceIncVAT"`
	Color              string       `json:"color"`
	Gearbox            string       `json:"gearbox"`
	Fuelstr            string       `json:"Fuelstr"`
	Body               string       `json:"body"`
	Power              int          `json:"power"`
	NewUsed            string       `json:"NewUsed"`
	SpecialPriceIncVAT float64      `json:"specialPriceIncVAT"`
	Vatornot           string       `json:"vatornot"`
	SellerId           string       `json:"sellerid"`
	ChannelCenshare    []Channels   `json:"channelscenshare"`
	OtherEquipment     string       `json:"othereq"`
	EquipmentList      []Equipments `json:"eqlist"`
	AdContent          string       `json:"adcontent"`
}
type Equipments struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type Channels struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type AllCarsData struct {
	Result []Car `json:"result"`
}

type BrandData struct {
	Result []Brand `json:"result"`
}

type ColorData struct {
	Full []Colors `json:"full"`
}

type BodyData struct {
	Full []Bodys `json:"full"`
}

type BlocketSite struct {
	Result []Sites `json:"result"`
}

type SellerData struct {
	Result []Sellers `json:"result"`
}
type Sites struct {
	Name      string   `json:"name"`
	SiteId    []string `json:"siteId"`
	BlocketId string   `json:"blocketid"`
}

type Colors struct {
	Sorting        int               `json:"sorting"`
	Name           string            `json:"name"`
	Value          string            `json:"value"`
	Localized_name map[string]string `json:"localized_name"`
}

type Bodys struct {
	Sorting        int               `json:"sorting"`
	Name           string            `json:"name"`
	Value          string            `json:"value"`
	Localized_name map[string]string `json:"localized_name"`
}

type Sellers struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

// Body
func findBody(wg *sync.WaitGroup) (body map[string]string, err error) {
	defer wg.Done()
	resp, err := resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/schema/priceBoard/cs-values/body")
	bodys := BodyData{}
	json.Unmarshal([]byte(resp.String()), &bodys)

	bodymap := make(map[string]string)
	for _, body := range bodys.Full {
		bodymap[body.Value] = body.Name
	}

	body = bodymap

	return body, nil

}

// Change VALUE of color to NAME
func findColor(wg *sync.WaitGroup) (color map[string]string, err error) {
	defer wg.Done()
	resp, err := resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/schema/priceBoard/cs-values/color")
	colors := ColorData{}
	json.Unmarshal([]byte(resp.String()), &colors)

	colormap := make(map[string]string)
	for _, color := range colors.Full {

		colormap[color.Value] = color.Name
	}
	color = colormap

	return color, nil
}

// Find sellers EMAIL by ID
func findSeller(wg *sync.WaitGroup) (sellers map[int]string, err error) {
	defer wg.Done()
	resp, err := resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/entity/sellerinfo?query=email&limit=2000")
	sellersdata := SellerData{}
	err = json.Unmarshal([]byte(resp.String()), &sellersdata)

	if err != nil {
		return nil, err
	}

	sellersmap := make(map[int]string)
	for _, seller := range sellersdata.Result {

		sellersmap[seller.Id] = seller.Email
	}

	sellers = sellersmap

	return sellers, nil
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var wg sync.WaitGroup

	wg.Add(3)
	// Sites

	var mysellers map[int]string
	var mycolors map[string]string
	var mybody map[string]string
	var err error

	go func() {
		mysellers, err = findSeller(&wg)
		mycolors, err = findColor(&wg)
		mybody, err = findBody(&wg)

		if err != nil {
			fmt.Println("here")
			fmt.Println(err)
		}

	}()

	wg.Wait()

	resp, err := resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/entity/blocketexport?query=blocketid&limit=1000")
	sites := BlocketSite{}
	json.Unmarshal([]byte(resp.String()), &sites)

	for _, site := range sites.Result {
		urlstr := "("
		for _, id := range site.SiteId {
			urlstr = urlstr + `site="` + id + `"` + `|`

		}
		urlstr = strings.TrimRight(urlstr, "|")
		urlstr = urlstr + ")"

		urlstr = urlstr + `%26(typeStr="PB"|typeStr="TR")%26(Wfstep=800)&site&limit=5000`
	

		fmt.Println(site.Name)
		fmt.Println("IAM VISTING %s", "http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/entity/priceBoard?query=Wfstep%3D800"+urlstr)

		// Creating the CSV
		resp, err = resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/entity/priceBoard?query=" + urlstr)
		if err != nil {
			fmt.Println(err)
		}

		//fmt.Println(resp.String())
		hedincars := AllCarsData{}
		json.Unmarshal([]byte(resp.String()), &hedincars)
		file, err := os.Create("/tmp/" + site.BlocketId + ".csv")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		// enc:=charmap.ISO8859_1.NewEncoder()

		csvWriter := csv.NewWriter(file)
		csvWriter.Comma = ';'

		for _, hedincar := range hedincars.Result {
			//fmt.Println(hedincar)
			goahead := 0
			for _, mycar := range hedincar.ChannelCenshare {
				//fmt.Println(mycar)

				if mycar.Key == "blocket" && mycar.Val == "1" {
					goahead = 1
				}
			}

			if goahead == 1 {

				resp, err = resty.R().Get("http://hedinbil-headless-sat-lb-prod-1244694856.eu-north-1.elb.amazonaws.com/hcms/v1.3/entity/brandlist?query=bid%3D%22" + hedincar.Brand + "%22")
				BrandFullName := BrandData{}
				json.Unmarshal([]byte(resp.String()), &BrandFullName)

				//	fmt.Println(hedincar.Brand)
				//	fmt.Println(hedincar.Regno)

				if len(BrandFullName.Result) <= 0 {
					fmt.Println("NO BRAND DATA")
					continue

				}
				hedincar.Brand = BrandFullName.Result[0].Name

				if hedincar.Brand == "DS Automobiles" {
					hedincar.Brand = "DS"
				}

				switch {
				case hedincar.Brand == "Ram Trucks":
					hedincar.Brand = "Dodge"

				case hedincar.Brand == "Rambler":
					hedincar.Brand = "Dodge"

				}

				//change fuel

				switch {
				case hedincar.Fuelstr == "Laddhybrid":
					hedincar.Fuelstr = "Hybrid el/bensin"

				case hedincar.Fuelstr == "Elhybrid":
					hedincar.Fuelstr = "Hybrid el/bensin"

				case hedincar.Fuelstr == "Övriga":
					hedincar.Fuelstr = "Inget"

				}

				// Change value from true/false to 1/0
				switch {

				case hedincar.Vatornot == "true":
					hedincar.Vatornot = "1"
					break

				case hedincar.Vatornot == "false":
					hedincar.Vatornot = "0"
					break

				case hedincar.Vatornot == "":
					hedincar.Vatornot = "0"
					break
				}

				// Change value of Gearbox from ENG to SWE
				switch {

				case hedincar.Gearbox == "Automatic":
					hedincar.Gearbox = "Automatisk"
					break

				case hedincar.Gearbox == "Manual":
					hedincar.Gearbox = "Manuell"
					break

				}

				// Change value of vehicletype from string to int
				switch {

				case hedincar.NewUsed == "NYTT":
					hedincar.NewUsed = "1"
					break

				case hedincar.NewUsed == "BEG":
					hedincar.NewUsed = "0"
					break

				case hedincar.NewUsed == "DEMO":
					hedincar.NewUsed = "0"
					break

				}

				if hedincar.Body == "Pickup" {
					hedincar.Body = "Flak"
				}

				if hedincar.Body == "Coupé" {
					hedincar.Body = "SportKupé"
				}

				

				switch hedincar.Body {

				case "Cabriolet":
					hedincar.Body = "Cab"
					break

				case "Grand Coupé":
					hedincar.Body = "SportKupé"
					break

				case "Coupé":
					hedincar.Body = "SportKupé"
					break

				case "Kombi-sedan":
					hedincar.Body = "Halvkombi"
					break

				case "pickup":
					hedincar.Body = "Flak"
					break

				case "minivan":
					hedincar.Body = "Minibuss"
					break

				case "van":
					hedincar.Body = "Skåp"
					break

				case "skap":
					hedincar.Body = "Skåp"
					break

				}

				fmt.Println("Iam NOW BODY!!! %s", hedincar.Body)

				// Transform model of MERCEDES, BMW
				switch {
				case hedincar.ModelManual == "A-Klass":
					hedincar.ModelManual = "A"

				case hedincar.ModelManual == "B-Klass":
					hedincar.ModelManual = "B"

				case hedincar.ModelManual == "C-Klass":
					hedincar.ModelManual = "C"

				case hedincar.ModelManual == "E-Klass":
					hedincar.ModelManual = "E"

				case hedincar.ModelManual == "S-Klass":
					hedincar.ModelManual = "S"

				case hedincar.ModelManual == "X-Klass":
					hedincar.ModelManual = "X"

				case hedincar.ModelManual == "V-Klass":
					hedincar.ModelManual = "V"

				case hedincar.ModelManual == "G-Klass":
					hedincar.ModelManual = "G"

				case hedincar.ModelManual == "1-serie":
					hedincar.ModelManual = "1"

				case hedincar.ModelManual == "2-serie":
					hedincar.ModelManual = "2"

				case hedincar.ModelManual == "3-serie":
					hedincar.ModelManual = "3"

				case hedincar.ModelManual == "4-serie":
					hedincar.ModelManual = "4"

				case hedincar.ModelManual == "5-serie":
					hedincar.ModelManual = "5"

				case hedincar.ModelManual == "6-serie":
					hedincar.ModelManual = "6"

				case hedincar.ModelManual == "7-serie":
					hedincar.ModelManual = "7"

				case hedincar.ModelManual == "8-serie":
					hedincar.ModelManual = "8"

				}

				// Placeholder for enviromental vehicle
				envVehicle := "0"

				var eqtot []string
				for _, myeq := range hedincar.EquipmentList {
					//fmt.Println(mycar)

					if myeq.Key == "equipment" {

						eqtot = append(eqtot, strings.TrimSpace(myeq.Val))
					}
				}

				eqSliceString := strings.Join(eqtot, ",")

				eqSliceString += "," + strings.TrimSpace(strings.TrimLeft(hedincar.OtherEquipment, ","))

				eqSliceString += hedincar.AdContent

				eqSliceString += hedincar.Regno

				eqSliceString = strings.Trim(eqSliceString, "\n")

				eqSliceString = strings.Trim(eqSliceString, "\r")

				eqSliceString = strings.Replace(eqSliceString, "\n", "", -1)

				eqSliceString = strings.Replace(eqSliceString, "\r", "", -1)

				eqSliceString = strings.Replace(eqSliceString, "\r\n", "", -1)

				sellerIdInt, _ := strconv.Atoi(hedincar.SellerId)

				salespriceString := strconv.FormatFloat(hedincar.SalesPriceIncVAT, 'f', -1, 64)
				specialpriceString := strconv.FormatFloat(hedincar.SpecialPriceIncVAT, 'f', -1, 64)
				x := []string{hedincar.Chassino, hedincar.Regno, hedincar.Brand, hedincar.ModelManual, hedincar.Modeltext, strconv.Itoa(hedincar.RegistrationDate), hedincar.Yearmodel,
					strconv.Itoa(hedincar.Mileage), salespriceString, mycolors[hedincar.Color], hedincar.Gearbox, hedincar.Fuelstr, eqSliceString, hedincar.Body, strconv.Itoa(hedincar.Power),
					hedincar.NewUsed, envVehicle, mysellers[sellerIdInt], specialpriceString,
					hedincar.Vatornot}
				csvWriter.Write(x)

			}
		}
		//end of results
		csvWriter.Flush()
	}
	//end of sites

	cfg, _ := external.LoadDefaultAWSConfig(
		// Specify the shared configuration profile to load.
		external.WithSharedConfigProfile("pontus.s.eclipse"),
	)

	cfg.Region = endpoints.EuWest1RegionID
	uploader := s3manager.NewUploader(cfg)

	var files []string

	root := "/tmp"
	err = filepath.Walk(root, func(fpath string, info os.FileInfo, err error) error {

		if path.Ext(fpath) == ".csv" {

			app := "/usr/bin/iconv"
			//arg0:="–-from-code=UTF-8 –-to-code=ISO-8859-1 "+fpath+" > "+fpath

			cmd := exec.Command(app, "-c", "-f", "UTF-8", "-t", "ISO-8859-1", fpath, "-o", fpath+".iso")
			err = cmd.Run()

			if err != nil {

				fmt.Println(err)
			}

			/*	cmd = exec.Command("/bin/sh", "-c", "mv "+fpath+".iso > "+fpath)

				stdout, err = cmd.Output()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(stdout)
			*/

			files = append(files, fpath)

		}

		return nil
	})

	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)

		fs3, err := os.OpenFile(file+".iso", os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}

		realpath := strings.Split(file, "/")
		// Upload the file to S3.
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String("hedin-censhare-prod-interfaces-bucket"),
			Key:    aws.String("blocket-export/" + realpath[2]),
			Body:   fs3,
		})
		if err != nil {
			fmt.Printf("failed to upload file, %v", err)

		}

		if err := fs3.Close(); err != nil {
			log.Fatal(err)
		}

	}

	return events.APIGatewayProxyResponse{
		Body:       "OK",
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
