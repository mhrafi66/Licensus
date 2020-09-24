package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Car :  Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Car struct {
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}

type SoloAsset struct {
	Holder string `json:"holder"`
	Level  string `json:"level"`
	Desc   string `json:"desc"`
}

type Resp struct {
	YourKey string `json:"yourkey"`
	Message string `json:"message"`
}

type License struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	NID    string `json:"nid"`
	Status string `json:"status"`
	Test1  string `json:"test1"`
	Test2  string `json:"test2"`
	Test3  string `json:"test3"`
	Point  string `json:"point"`
}

type TrafficRuleViolatonReport struct {
	ID              string `json:"id"`
	Holder          string `json:"holder"`
	Level           string `json:"level"`
	Desc            string `json:"desc"`
	PointsDeduction string `json:"pointsdeduction"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("licensus_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()

	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	if function == "queryLicense" {
		return s.queryLicense(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createLearnerLicense" {
		return s.createLearnerLicense(APIstub, args)
	} else if function == "inputTest1Result" {
		return s.inputTest1Result(APIstub, args)
	} else if function == "inputTest2Result" {
		return s.inputTest2Result(APIstub, args)
	} else if function == "inputTest3Result" {
		return s.inputTest3Result(APIstub, args)
	} else if function == "getHistoryForAsset" {
		return s.getHistoryForAsset(APIstub, args)
	} else if function == "queryLearnerList" {
		return s.queryLearnerList(APIstub, args)
	} else if function == "restictedMethod" {
		return s.restictedMethod(APIstub, args)
	} else if function == "queryWaitingList" {
		return s.queryWaitingList(APIstub, args)
	} else if function == "upgradeLearnerToActive" {
		return s.upgradeLearnerToActive(APIstub, args)
	} else if function == "queryActiveList" {
		return s.queryActiveList(APIstub, args)
	} else if function == "createPoliceReport" {
		return s.createPoliceReport(APIstub, args)
	} else if function == "queryComplainByLicenseNo" {
		return s.queryComplainByLicenseNo(APIstub, args)
	} else if function == "queryToStallList" {
		return s.queryToStallList(APIstub, args)
	} else if function == "revokeLicense" {
		return s.revokeLicense(APIstub, args)
	} else if function == "queryStalledList" {
		return s.queryStalledList(APIstub, args)
	} else if function == "deleteLicense" {
		return s.deleteLicense(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (S *SmartContract) deleteLicense(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-Approver have access this method!")
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	current := "current"
	colorNameIndexKey, err := APIstub.CreateCompositeKey("learner~key", []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	APIstub.DelState(colorNameIndexKey)

	colorNameIndexKey, err = APIstub.CreateCompositeKey("active~key", []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	APIstub.DelState(colorNameIndexKey)

	colorNameIndexKey, err = APIstub.CreateCompositeKey("stalled~key", []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	APIstub.DelState(colorNameIndexKey)

	APIstub.DelState(args[0])

	return shim.Success(nil)
}

func (S *SmartContract) queryStalledList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	current := "current"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("stalled~key", []string{current})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		// frac := "," + id

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, newBytes...)

		} else {
			// newBytes := append([]byte(id), assetAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (S *SmartContract) queryToStallList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" && val != "org2-police" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1 have access this method!")
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	current := "current"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("tostall~key", []string{current})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		// frac := "," + id

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, newBytes...)

		} else {
			// newBytes := append([]byte(id), assetAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (s *SmartContract) revokeLicense(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-Approver have access this method!")
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	indexName := "active~key"
	current := "current"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	APIstub.DelState(colorNameIndexKey)

	indexName = "tostall~key"
	current = "current"
	colorNameIndexKey, err = APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	APIstub.DelState(colorNameIndexKey)

	license.Status = "Stalled"

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	indexName = "stalled~key"

	colorIndexKey, errr := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if errr != nil {
		return shim.Error(errr.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorIndexKey, value)

	return shim.Success(licenseAsBytes)
}

func (S *SmartContract) queryComplainByLicenseNo(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	licenseNo := args[0]

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("crime~key", []string{licenseNo})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			appendBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, appendBytes...)

		} else {
			// newBytes := append([]byte(","), carsAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (S *SmartContract) queryActiveList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	current := "current"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("active~key", []string{current})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		// frac := "," + id

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, newBytes...)

		} else {
			// newBytes := append([]byte(id), assetAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (s *SmartContract) createPoliceReport(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org2-police" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG2-Police have access this method!")
	}

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var trv = TrafficRuleViolatonReport{ID: args[0], Holder: args[1], Level: args[2], Desc: args[3], PointsDeduction: args[4]}

	trvAsBytes, _ := json.Marshal(trv)
	APIstub.PutState(args[0], trvAsBytes)

	indexName := "crime~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{args[1], args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	pdeduce, errp := strconv.Atoi(args[4])
	if errp != nil {
		return shim.Error(errp.Error())
	}

	licenseAsBytes, _ := APIstub.GetState(args[1])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	premaining, errp2 := strconv.Atoi(license.Point)
	if errp2 != nil {
		return shim.Error(errp2.Error())
	}

	pupdated := premaining - pdeduce

	license.Point = strconv.Itoa(pupdated)

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[1], licenseAsBytes)

	if pupdated <= 0 {
		indexName5 := "tostall~key"
		current5 := "current"
		colorNameIndexKey5, err5 := APIstub.CreateCompositeKey(indexName5, []string{current5, args[1]})
		if err5 != nil {
			return shim.Error(err5.Error())
		}
		value5 := []byte{0x00}
		APIstub.PutState(colorNameIndexKey5, value5)
	}

	return shim.Success(trvAsBytes)
}

func (S *SmartContract) queryWaitingList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-Approver have access this method!")
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	current := "current"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("waiting~key", []string{current})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		// frac := "," + id

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, newBytes...)

		} else {
			// newBytes := append([]byte(id), assetAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (s *SmartContract) upgradeLearnerToActive(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-Approver have access this method!")
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	indexName := "learner~key"
	current := "current"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	APIstub.DelState(colorNameIndexKey)

	indexName = "waiting~key"
	current = "current"
	colorNameIndexKey, err = APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	APIstub.DelState(colorNameIndexKey)

	license.Status = "Active"

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	indexName = "active~key"

	colorIndexKey, errr := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if errr != nil {
		return shim.Error(errr.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorIndexKey, value)

	return shim.Success(licenseAsBytes)
}

func (s *SmartContract) queryLicense(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(licenseAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	licenses := []License{
		License{ID: "LICENSE0", Name: "Test1", NID: "123", Status: "Active", Test1: "Yes", Test2: "Yes", Test3: "Yes", Point: "15"},
		License{ID: "LICENSE1", Name: "Test1", NID: "123", Status: "Active", Test1: "Yes", Test2: "Yes", Test3: "Yes", Point: "15"},
	}

	i := 0
	for i < len(licenses) {
		licenseAsBytes, _ := json.Marshal(licenses[i])
		APIstub.PutState("LICENSE"+strconv.Itoa(i), licenseAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createLearnerLicense(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-Approver have access this method!")
	}

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	keyExists, err := APIstub.GetState(args[0])
	if keyExists != nil {
		return shim.Error("Key already exists")
	}

	nidExist, errn := APIstub.CreateCompositeKey("nid~key", []string{"current", args[2]})
	if errn != nil {
		return shim.Error(errn.Error())
	}

	keyExists, err = APIstub.GetState(nidExist)
	if keyExists != nil {
		return shim.Error("NID already exists")
	}

	var license = License{ID: args[0], Name: args[1], NID: args[2], Status: "Learner", Test1: "No", Test2: "No", Test3: "No", Point: "15"}

	licenseAsBytes, _ := json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	indexName := "learner~key"
	current := "current"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)
	APIstub.PutState(nidExist, value)

	return shim.Success(licenseAsBytes)
}

func (S *SmartContract) queryLearnerList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	current := "current"

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("learner~key", []string{current})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var licenses []byte
	bArrayMemberAlreadyWritten := false

	licenses = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		// frac := "," + id

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			licenses = append(licenses, newBytes...)

		} else {
			// newBytes := append([]byte(id), assetAsBytes...)
			licenses = append(licenses, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	licenses = append(licenses, []byte("]")...)

	return shim.Success(licenses)
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// get an ID for the client which is guaranteed to be unique within the MSP
	//id, err := cid.GetID(APIstub) -

	// get the MSP ID of the client's identity
	//mspid, err := cid.GetMSPID(APIstub) -

	// get the value of the attribute
	//val, ok, err := cid.GetAttributeValue(APIstub, "attr1") -

	// get the X509 certificate of the client, or nil if the client's identity was not based on an X509 certificate
	//cert, err := cid.GetX509Certificate(APIstub) -

	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1 have access this method!")
	}
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(carAsBytes)

}

func (s *SmartContract) inputTest1Result(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-examcenter1" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-ExamCenter1 have access this method!")
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	license.Test1 = args[1]

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	if license.Test1 == "Yes" && license.Test2 == "Yes" && license.Test3 == "Yes" {
		indexName := "waiting~key"
		current := "current"
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
		if err != nil {
			return shim.Error(err.Error())
		}
		value := []byte{0x00}
		APIstub.PutState(colorNameIndexKey, value)
	}

	return shim.Success(licenseAsBytes)
}

func (s *SmartContract) inputTest2Result(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-examcenter2" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-ExamCenter2 have access this method!")
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	license.Test2 = args[1]

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	if license.Test1 == "Yes" && license.Test2 == "Yes" && license.Test3 == "Yes" {
		indexName := "waiting~key"
		current := "current"
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
		if err != nil {
			return shim.Error(err.Error())
		}
		value := []byte{0x00}
		APIstub.PutState(colorNameIndexKey, value)
	}

	return shim.Success(licenseAsBytes)
}

func (s *SmartContract) inputTest3Result(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "org1-examcenter3" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as ORG1-ExamCenter3 have access this method!")
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	licenseAsBytes, _ := APIstub.GetState(args[0])
	license := License{}

	json.Unmarshal(licenseAsBytes, &license)

	license.Test3 = args[1]

	licenseAsBytes, _ = json.Marshal(license)
	APIstub.PutState(args[0], licenseAsBytes)

	if license.Test1 == "Yes" && license.Test2 == "Yes" && license.Test3 == "Yes" {
		indexName := "waiting~key"
		current := "current"
		colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{current, args[0]})
		if err != nil {
			return shim.Error(err.Error())
		}
		value := []byte{0x00}
		APIstub.PutState(colorNameIndexKey, value)
	}

	return shim.Success(licenseAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	licenseID := args[0]

	resultsIterator, err := stub.GetHistoryForKey(licenseID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
