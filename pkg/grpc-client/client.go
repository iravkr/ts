
package main

import (
	// "fmt"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/onosproject/cco-mon/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port     = ":5150"
filePath = "model-cells-10sec-more-ues.csv"
)

// variables to store and update cell info data
var cellInfo = make(map[string]float32)
var cgistring []string

type OcnRow struct {
	Column string
	Value  string
}

type Column string
type Cgi string

func main() {

	// Replace with the IP Address of the server running K8s node
	conn, err := grpc.Dial("192.168.238.206"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	//delete the csv file contents
	if err := os.Truncate(filePath, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	//new grpc client
	client := pb.NewCCOMonitoringServiceClient(conn)

	//TODO: error handling
	callGetCellInfo(client)

	// fmt.Println("Fething ocns...")
	resp := callGetOcn(client)
	// log.Println("the ocn map is %v", resp) 
	for key := range cellInfo {
		cgistring = append(cgistring, key)
	}
	columns := map[Column]bool{}
	ocnMap := map[Cgi]map[Column]string{}
	//rows := []OcnRow{}
	for key, value := range resp {
		cgis := strings.Split(key, ":")
		// log.Println("the cgis is %v", cgis) 

		cgi := cgis[len(cgis)-1]
		// log.Println("the cgi is %v", cgi) 

		ocnMap[Cgi(cgi)] = map[Column]string{}
		for ocnKey, value := range value.OcnRecord {
			parts := strings.Split(ocnKey, ":")
			if len(parts) >= 3 {
				// Concatenate with a '0' in the middle as per your requirement
				column := parts[1] + "0" + parts[2]
				columns[Column(column)] = true
				ocnMap[Cgi(cgi)][Column(column)] = fmt.Sprintf("%d", value)
			}
		}
	}

	// fmt.Println(ocnMap) // we need to fix
	quit := make(chan bool)

	//go func for collecting rsrp reports
	go func() {
		for {
			select {
			default:
				// log.Println("the ocn map is %v", resp) 
				log.Println("RSRP reports")
				time.Sleep(10 * time.Second)
				//TODO: error handling
				// initMongoDB()
				// clearMongoDBCollection()
				callGetRsrpReports(client, columns, ocnMap)
				//callGetRsrpReports(client)

			case <-quit:
				log.Print("Quitting the programm now")
				return
			}

		}
	}()

	time.Sleep(36000 * time.Second)
	quit <- true
}



// package main

// import (
// 	// "fmt"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strings"
// 	"time"

// 	pb "github.com/onosproject/cco-mon/pkg/proto"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// const (
// 	port     = ":5150"
// filePath = "model-cells-10sec-more-ues.csv"
// )

// // variables to store and update cell info data
// var cellInfo = make(map[string]float32)
// var cgistring []string

// type OcnRow struct {
// 	Column string
// 	Value  string
// }

// type Column string
// type Cgi string

// func main() {

// 	// Replace with the IP Address of the server running K8s node
// 	conn, err := grpc.Dial("192.168.184.144"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("Did not connect: %v", err)
// 	}
// 	defer conn.Close()

// 	//delete the csv file contents
// 	if err := os.Truncate(filePath, 0); err != nil {
// 		log.Printf("Failed to truncate: %v", err)
// 	}

// 	//new grpc client
// 	client := pb.NewCCOMonitoringServiceClient(conn)

// 	//TODO: error handling
// 	callGetCellInfo(client)

// 	// // fmt.Println("Fething ocns...")
// 	// resp := callGetOcn(client)
// 	// // log.Println("the ocn map is %v", resp) 
// 	// for key := range cellInfo {
// 	// 	cgistring = append(cgistring, key)
// 	// }
// 	columns := map[Column]bool{}
// 	ocnMap := map[Cgi]map[Column]string{}
// 	// //rows := []OcnRow{}
// 	// for key, value := range resp {
// 	// 	cgis := strings.Split(key, ":")
// 	// 	// log.Println("the cgis is %v", cgis) 

// 	// 	cgi := cgis[len(cgis)-1]
// 	// 	// log.Println("the cgi is %v", cgi) 

// 	// 	ocnMap[Cgi(cgi)] = map[Column]string{}
// 	// 	for ocnKey, value := range value.OcnRecord {
// 	// 		parts := strings.Split(ocnKey, ":")
// 	// 		if len(parts) >= 3 {
// 	// 			// Concatenate with a '0' in the middle as per your requirement
// 	// 			column := parts[1] + "0" + parts[2]
// 	// 			columns[Column(column)] = true
// 	// 			ocnMap[Cgi(cgi)][Column(column)] = fmt.Sprintf("%d", value)
// 	// 		}
// 	// 	}
// 	// }

// 	// fmt.Println(ocnMap) // we need to fix
// 	quit := make(chan bool)

// 	//go func for collecting rsrp reports

// 	//go func for collecting rsrp reports
// go func() {
//     for {
//         select {
//         default:
//             log.Println("RSRP reports")
//             time.Sleep(2 * time.Second)

//             // Fetch OCN data every iteration
//             resp := callGetOcn(client)  // Move this line inside the loop
//             for key, value := range resp {
//                 cgis := strings.Split(key, ":")
//                 cgi := cgis[len(cgis)-1]
//                 ocnMap[Cgi(cgi)] = map[Column]string{}
//                 for ocnKey, value := range value.OcnRecord {
//                     parts := strings.Split(ocnKey, ":")
//                     if len(parts) >= 3 {
//                         column := parts[1] + "0" + parts[2]
//                         columns[Column(column)] = true
//                         ocnMap[Cgi(cgi)][Column(column)] = fmt.Sprintf("%d", value)
//                     }
//                 }
//             }

//             // Call to get new RSRP reports with updated OCN data
//             callGetRsrpReports(client, columns, ocnMap)

//         case <-quit:
//             log.Print("Quitting the program now")
//             return
//         }
//     }
// }()

// time.Sleep(36000 * time.Second)
// quit <- true

// 	// go func() {
// 	// 	for {
// 	// 		select {
// 	// 		default:
// 	// 			// log.Println("the ocn map is %v", resp) 
// 	// 			log.Println("RSRP reports")
// 	// 			time.Sleep(10 * time.Second)
// 	// 			//TODO: error handling
// 	// 			// initMongoDB()
// 	// 			// clearMongoDBCollection()
// 	// 			callGetRsrpReports(client, columns, ocnMap)
// 	// 			//callGetRsrpReports(client)

// 	// 		case <-quit:
// 	// 			log.Print("Quitting the programm now")
// 	// 			return
// 	// 		}

// 	// 	}
// 	// }()
// 	// go func() {
// 	// 	for {
// 	// 		log.Println("RSRP reports")
// 	// 		time.Sleep(5 * time.Second)
	
// 	// 		// Directly call the function since it handles its errors internally and does not terminate the program
// 	// 		callGetRsrpReports(client, columns, ocnMap)
	
// 	// 		// No need to handle errors here since `callGetRsrpReports` does not return any
// 	// 		// If implementing retry logic, it should be encapsulated within `callGetRsrpReports` itself,
// 	// 		// or you would need to change its signature to return an error.
// 	// 	}
// 	// }()
	


// 	// go func() {
// 	// 	for {
// 	// 		// Attempt to collect RSRP reports.
// 	// 		// If an error occurs, log it and retry after a delay.
// 	// 		err := callGetRsrpReports(client, columns, ocnMap)
// 	// 		if err != nil {
// 	// 			log.Printf("Error collecting RSRP reports: %v. Retrying in 5 seconds...", err)
// 	// 			time.Sleep(5 * time.Second)
// 	// 			continue
// 	// 		}
	
// 	// 		// If the function completes without error, you may choose to exit the loop
// 	// 		// or implement some condition for it to run again.
// 	// 		// Here, we'll simply log and wait before restarting the process.
// 	// 		log.Println("Completed collecting RSRP reports. Restarting after 5 seconds...")
// 	// 		time.Sleep(5 * time.Second)
// 	// 	}
// 	// }()
	


// 	/* LOGIC FOR CHANGING TRX POWER FOR CELLS */
// 	// key := "13842601455c001"
// 	// time.Sleep(300 * time.Second)

// 	// log.Printf("Setting the ptx power of cell %v from %v to %v db", key, cellInfo[key], 33)
// 	//TODO: error handling
// 	// res := callSetCellPTX(client, key, float32(33))
// 	// if res == "Updated" {
// 	// 	log.Printf("Cell data updated: Cell %v, new Tx Power %v", key, cellInfo[key])
// 	// } else {
// 	// 	log.Println("Cell ptx update was not successful")
// 	// }

// 	// time.Sleep(300 * time.Second)
// 	// log.Printf("Setting the ptx power t cell %v to %v db", key, 11)
// 	//TODO: error handling
// 	// res = callSetCellPTX(client, key, float32(11))
// 	// if res == "Updated" {
// 	// 	log.Printf("Cell data updated: Cell %v, new Tx Power %v", key, cellInfo[key])
// 	// } else {
// 	// 	log.Println("Cell ptx update was not successful")
// 	// }

// 	// time.Sleep(36000 * time.Second)
// 	// log.Printf("Setting the ptx power t cell %v to %v db", key, -11)
// 	//TODO: error handling
// 	// res = callSetCellPTX(client, key, float32(-11))
// 	// if res == "Updated" {
// 	// 	log.Printf("Cell data updated: Cell %v, new Tx Power %v", key, cellInfo[key])
// 	// } else {
// 	// 	log.Println("Cell ptx update was not successful")
// 	// }

// 	// time.Sleep(300 * time.Second)
// 	// log.Printf("Setting the ptx power t cell %v to %v db", key, 11)
// 	// //TODO: error handling
// 	// // res = callSetCellPTX(client, key, float32(11))
// 	// // if res == "Updated" {
// 	// // 	log.Printf("Cell data updated: Cell %v, new Tx Power %v", key, cellInfo[key])
// 	// // } else {
// 	// // 	log.Println("Cell ptx update was not successful")
// 	// // }

// 	// time.Sleep(300 * time.Second)

// 	// Quit the program
// 	// quit <- true
// // for i := 20; i > -30; i -= 20 {
// // 		time.Sleep(100 * time.Second)
// // 		log.Printf("Setting the ptx power t cell %v to %v db", key, i)
// // 		res := callSetCellPTX(client, key, float32(i))
// // 		if res == "Updated" {
// // 			log.Printf("Cell data updated: Cell %v, new Tx Power %v", key, cellInfo[key])
// // 		} else {
// // 			log.Println("Cell ptx update was not successful")
// // 		}
// // 	}
// }
