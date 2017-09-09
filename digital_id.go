package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type DocumentInfo struct {
	Owner string   `json:"owner"`
	Docs  []string `json:"docs"`
}
type User struct {
	Owns []string `json:"owns"`
	//SharedwithMe []DocumentInfo `json:"sharedwithme"`
	SharedwithMe map[string][]string `json:"sharedwithme"`
	Auditrail    map[string][]string `json:"audittrail"`
}

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "getMydocs":
		return t.getMydocs(stub, args)
	case "getSharedDocs":
		return t.getMydocs(stub, args)

	}
	return nil, nil
}
*/
func (t *SimpleChaincode) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	switch function {
	case "createUser":
		return t.createUser(APIstub, args)
	case "addDocument":
		return t.addDocument(APIstub, args)
	case "shareDocument":
		return t.shareDocument(APIstub, args)
	case "revokeAccess":
		return t.revokeAccess(APIstub, args)
	case "removeDocument":
		return t.removeDocument(APIstub, args)
	case "getMydocs":
		return t.getMydocs(APIstub, args)
	case "getSharedDocs":
		return t.getMydocs(APIstub, args)

	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (t *SimpleChaincode) removeDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 2 {
		fmt.Println("Expecting a minimum of three arguments Argument")
		//return nil, errors.New("Expected at least one arguments for adding a user")
		return shim.Error("error")
	}

	var userhash = args[0]
	var dochash = args[1]
	var user User
	var err error
	user, err = readFromBlockchain(userhash, APIstub)
	if err != nil {
		//	return nil, errors.New("failed to read")
		return shim.Error("error")
	}

	for i, v := range user.Owns {
		if v == dochash {
			user.Owns = append(user.Owns[:i], user.Owns[i+1:]...)
			break
		}
	}

	_, err = writeIntoBlockchain(userhash, user, APIstub)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		//return nil, err
		return shim.Error("error")
	}

	fmt.Println("Successfully removed the doc")
	return shim.Success(nil)

}

func (t *SimpleChaincode) revokeAccess(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 3 {
		fmt.Println("Expecting a minimum of three arguments Argument")
		//return nil, errors.New("Expected at least one arguments for adding a user")
		return shim.Error("error")
	}

	var userhash = args[0]
	var orghash = args[1]
	var dochash = args[2]

	var org User

	org, err := readFromBlockchain(orghash, APIstub)
	if err != nil {
		//return nil, errors.New("failed to read")
		return shim.Error("error")
	}

	userDocsArray := org.SharedwithMe[userhash]

	// removes that particular document from the array
	for i, v := range userDocsArray {
		if v == dochash {
			userDocsArray = append(userDocsArray[:i], userDocsArray[i+1:]...)
			break
		}
	}

	//assign that array to the user map key
	org.SharedwithMe[userhash] = userDocsArray

	_, err = writeIntoBlockchain(orghash, org, APIstub)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		return shim.Error("error")
	}

	fmt.Println("Successfully revoked access to the doc")
	return shim.Success(nil)

}
func (t *SimpleChaincode) createUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//func createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering createUser")

	if len(args) < 1 {
		fmt.Println("Expecting One Argument")
		//	return nil, errors.New("Expected at least one arguments for adding a user")
		return shim.Error("error")
	}

	var userid = args[0]
	var userinfo = `{"owns":[],"mymap":{}, "audit":{}}`

	err := APIstub.PutState(userid, []byte(userinfo))
	if err != nil {
		fmt.Println("Could not save user to ledger", err)
		//return nil, err
		return shim.Error("error")
	}

	fmt.Println("Successfully saved user/org")
	return shim.Success(nil)
}

//2.addDocument()   (#user,#doc)
func (t *SimpleChaincode) addDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("Entering addDocument")
	var user User
	if len(args) < 2 {
		fmt.Println("Expecting two Argument")
		//return nil, errors.New("Expected at least two arguments for adding a document")
		//return shim.Error(errors.New("Expected at least two arguments for adding a document"))
		return shim.Error("error")
	}

	var userid = args[0]
	fmt.Println(userid)
	var docid = args[1]
	fmt.Println(docid)
	bytes, err := APIstub.GetState(userid)
	if err != nil {
		//	return nil, err
		return shim.Error("error")
	}

	err = json.Unmarshal(bytes, &user)
	if err != nil {
		fmt.Println("unable to unmarshal user data")
		//return nil, err
		return shim.Error("error")
	}

	user.Owns = append(user.Owns, docid)

	_, err = writeIntoBlockchain(userid, user, APIstub)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		//return nil, err
		return shim.Error("error")
	}

	fmt.Println("Successfully added the doc")
	return shim.Success(nil)

}

//3. shareDocument()    (#doc,#user, #org)  Invoke
func (t *SimpleChaincode) shareDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("Entering shareDocument")
	var user User
	var org User
	//	var doc DocumentInfo
	//fmt.Println(doc)
	if len(args) < 2 {
		fmt.Println("Expecting three Argument")
		//return nil, errors.New("Expected at least three arguments for sharing  a document")
		//return shim.Error(errors.New("Expected at least two arguments for adding a document"))
		return shim.Error("error")
	}

	var userid = args[0]
	var docid = args[1]
	var orgid = args[2]
	//fetching the user
	userbytes, err := APIstub.GetState(userid)
	if err != nil {
		fmt.Println("could not fetch user", err)
		//return nil, err
		return shim.Error("error")

	}
	err = json.Unmarshal(userbytes, &user)
	if err != nil {
		fmt.Println("unable to unmarshal user data")
		//return nil, err
		return shim.Error("error")
	}
	if !contains(user.Owns, docid) {
		fmt.Println("docment doesnt exists")
		// return nil, err
		return shim.Error("error")
	}
	//fetch oraganisation
	orgbytes, err := APIstub.GetState(orgid)
	if err != nil {
		fmt.Println("could not fetch user", err)
		// return nil, err
		return shim.Error("error")
	}
	err = json.Unmarshal(orgbytes, &org)
	if err != nil {
		fmt.Println("unable to unmarshal org data")
		// return nil, err
		return shim.Error("error")
	}

	if org.SharedwithMe == nil {
		org.SharedwithMe = make(map[string][]string)
	}

	if user.Auditrail == nil {
		user.Auditrail = make(map[string][]string)

	}
	//adding the document if it doesnt exists already
	if !contains(org.SharedwithMe[userid], docid) {
		timestamp := makeTimestamp()
		fmt.Println(timestamp)
		//---------------Sharing the doc to Organisation-----------------------
		org.SharedwithMe[userid] = append(org.SharedwithMe[userid], docid)

		//-------------- Adding time stamp to user audit trail array-------------
		user.Auditrail[orgid] = append(user.Auditrail[orgid], timestamp)
		user.Auditrail[orgid] = append(user.Auditrail[orgid], docid)
	}

	_, err = writeIntoBlockchain(orgid, org, APIstub)
	if err != nil {
		fmt.Println("Could not save org Info", err)
		// return nil, err
		return shim.Error("error")
	}

	_, err = writeIntoBlockchain(userid, user, APIstub)
	if err != nil {
		fmt.Println("Could not save user Info", err)
		// return nil, err
		return shim.Error("error")
	}

	fmt.Println("Successfully shared the doc")
	return shim.Success(nil)

}

//4. getMydocs()    (#user) Query
func (t *SimpleChaincode) getMydocs(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("Entering get my docs")

	if len(args) < 1 {
		fmt.Println("Invalid number of arguments")
		// return nil, errors.New("Missing userid")
		//return shim.Error(errors.New("Expected at least one arguments for fetching document i.e userid"))
		return shim.Error("error")
	}

	var userid = args[0]
	idasbytes, err := APIstub.GetState(userid)
	if err != nil {
		fmt.Println("Could not user info", err)
		// return nil, err
		return shim.Error("error")
	}
	//return idasbytes, nil
	return shim.Success(idasbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))

	if err != nil {
		fmt.Println("Could not start SimpleChaincode")
	} else {
		fmt.Println("SimpleChaincode successfully started")
	}
}
func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
func makeTimestamp() string {
	t := time.Now()

	return t.Format(("2006-01-02T15:04:05.999999-07:00"))
	//return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

//------------- reusable methods -------------------
func writeIntoBlockchain(key string, value User, APIstub shim.ChaincodeStubInterface) ([]byte, error) {
	bytes, err := json.Marshal(&value)
	if err != nil {
		fmt.Println("Could not marshal info object", err)
		return nil, err
	}

	err = APIstub.PutState(key, bytes)
	if err != nil {
		fmt.Println("Could not save sharing info to org", err)
		return nil, err
	}

	return nil, nil
}

func readFromBlockchain(key string, APIstub shim.ChaincodeStubInterface) (User, error) {
	userbytes, err := APIstub.GetState(key)
	var user User
	if err != nil {
		fmt.Println("could not fetch user", err)
		return user, err
	}

	err = json.Unmarshal(userbytes, &user)
	if err != nil {
		fmt.Println("Unable to marshal data", err)
		return user, err
	}

	return user, nil
}
