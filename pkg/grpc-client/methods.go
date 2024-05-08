package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	// "strings"
	// "math"
	// "sort"

	

	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/bson"
	// "time"
	//"reflect"

	pb "github.com/onosproject/cco-mon/pkg/proto"
)

const (
    Cvoice = 300.0 // in Kbps
    Cembb  = 3000.0 // in Kbps for 3 Mbps
    Bcell  = 20000.0 // Total bandwidth for each cell in Kbps (20 MHz = 20,000 Kbps)
)


// func callGetRsrpReports(client pb.CCOMonitoringServiceClient, columns map[Column]bool, ocnMap map[Cgi]map[Column]string) {
//     var array [][]string

//     stream, err := client.GetRsrpReports(context.Background(), &pb.NoParam{})
//     if err != nil {
//         log.Printf("Could not get rsrp_info_client: %v", err)
//         return
//     }

//     headers := []string{"Time", "UeID", "FiveQi", "S_CGI", "RsrpServing", "RsrpNeigbors"}

//     // Sort the map keys to maintain consistent column order
//     sortedColumns := make([]Column, 0, len(columns))
//     for column := range columns {
//         sortedColumns = append(sortedColumns, column)
//     }
//     sort.Slice(sortedColumns, func(i, j int) bool {
//         return sortedColumns[i] < sortedColumns[j]
//     })

//     for _, column := range sortedColumns {
//         headers = append(headers, string(column))
//     }
//     array = append(array, headers)

//     for {
//         info, err := stream.Recv()
//         if err == io.EOF {
//             break
//         }
//         if err != nil {
//             log.Printf("Error while streaming %v", err)
//             return
//         }

//         rsrpNeighbors := fmt.Sprint(info.RsrpNeighbors)[4 : len(fmt.Sprint(info.RsrpNeighbors))-1]
//         columnMap := ocnMap[Cgi(info.Cgi)]
//         row := []string{info.Time, info.Ueid, strconv.Itoa(int(info.Fiveqi)), info.Cgi, strconv.Itoa(int(info.RsrpServing)), rsrpNeighbors}
//         fmt.Printf("row %+v\n", row)

//         // Append values in the order of sorted columns
//         for _, column := range sortedColumns {
//             val, ok := columnMap[column]
//             if ok {
//                 row = append(row, val)
//             } else {
//                 row = append(row, "0")
//             }
//         }

//         array = append(array, row)
//     }

//     CSVFileWriter(array)
// }


/// ********************* Final version if you want Ueid list and UEthroughput of each UEId caculated ***************************


// type CellData struct {
//     AvgSIR          float64
//     Count5Qi        [2]int
//     RSRPServingDB   float64
//     AvgNeighborsDB  float64
//     DynamicValues   []string
//     UeIDsVoice      []string
//     UeIDsEMBB       []string
//     UEThroughputVoice []string
//     UEThroughputEMBB  []string
// }

// func callGetRsrpReports(client pb.CCOMonitoringServiceClient, columns map[Column]bool, ocnMap map[Cgi]map[Column]string) {
//     var array [][]string
//     sirMap := make(map[string][]float64)
//     cellData := make(map[string]*CellData)

//     stream, err := client.GetRsrpReports(context.Background(), &pb.NoParam{})
//     if err != nil {
//         log.Printf("Could not get rsrp_info_client: %v", err)
//         return
//     }

//     dynamicHeaders := make([]string, 0, len(columns))
//     for column := range columns {
//         dynamicHeaders = append(dynamicHeaders, string(column))
//     }
//     sort.Strings(dynamicHeaders)
//     headers := append([]string{"Serving ID", "rsrpServingDB", "avgNeighborsDB", "count5Qi[1]", "count5Qi[2]", "Bvoice", "Bembb", "Avg SIR", "BandwidthUsed", "BandwidthUnused"}, dynamicHeaders...)
//     headers = append(headers, "UeID-voice", "UeID-embb", "UEthroughput-voice", "UEthroughput-embb")
//     array = append(array, headers)

//     for {
//         info, err := stream.Recv()
//         if err == io.EOF {
//             break
//         }
//         if err != nil {
//             log.Printf("Error while streaming %v", err)
//             return
//         }

//         if _, ok := cellData[info.Cgi]; !ok {
//             cellData[info.Cgi] = &CellData{
//                 DynamicValues: make([]string, len(dynamicHeaders)),
//                 UeIDsVoice: make([]string, 0),
//                 UeIDsEMBB: make([]string, 0),
//                 UEThroughputVoice: make([]string, 0),
//                 UEThroughputEMBB: make([]string, 0),
//             }
//         }
//         data := cellData[info.Cgi]

//         rsrpNeighbors := fmt.Sprint(info.RsrpNeighbors)[4 : len(fmt.Sprint(info.RsrpNeighbors))-1]
//         neighbors := strings.Split(rsrpNeighbors, " ")
//         var linearSum float64
//         for _, neighbor := range neighbors {
//             rsrp := strings.Split(neighbor, ":")[1]
//             n, _ := strconv.Atoi(rsrp)
//             linear := math.Pow(10, float64(n)/10)
//             linearSum += linear
//         }
//         data.AvgNeighborsDB = 10 * math.Log10(linearSum / float64(len(neighbors)))
//         data.RSRPServingDB = float64(info.RsrpServing)
//         sir := data.RSRPServingDB - data.AvgNeighborsDB
//         sirMap[info.Cgi] = append(sirMap[info.Cgi], sir)

//         // Update counts and track UE IDs
//         if info.Fiveqi == 1 {
//             data.Count5Qi[0]++
//             data.UeIDsVoice = append(data.UeIDsVoice, info.Ueid)
//             throughput := Cvoice * math.Log2(1 + math.Pow(10, sir/10))
//             data.UEThroughputVoice = append(data.UEThroughputVoice, fmt.Sprintf("%s: %.2f", info.Ueid, throughput))
//         } else if info.Fiveqi == 2 {
//             data.Count5Qi[1]++
//             data.UeIDsEMBB = append(data.UeIDsEMBB, info.Ueid)
//             throughput := Cembb * math.Log2(1 + math.Pow(10, sir/10))
//             data.UEThroughputEMBB = append(data.UEThroughputEMBB, fmt.Sprintf("%s: %.2f", info.Ueid, throughput))
//         }

//         // Capture dynamic column values
//         columnMap := ocnMap[Cgi(info.Cgi)]
//         for idx, col := range dynamicHeaders {
//             val, ok := columnMap[Column(col)]
//             if ok {
//                 data.DynamicValues[idx] = val
//             } else {
//                 data.DynamicValues[idx] = "N/A"
//             }
//         }
//     }

//     // Output results
//     for id, data := range cellData {
//         total := 0.0
//         for _, s := range sirMap[id] {
//             total += s
//         }
//         sirAvg := total / float64(len(sirMap[id]))
//         Bvoice := Cvoice / math.Log2(1+math.Pow(10, sirAvg/10))
//         Bembb := Cembb / math.Log2(1+math.Pow(10, sirAvg/10))

//         BandwidthUsed := float64(data.Count5Qi[0])*Bvoice + float64(data.Count5Qi[1])*Bembb
//         BandwidthUnused := Bcell - BandwidthUsed

//         row := []string{
//             id,
//             fmt.Sprintf("%.2f", data.RSRPServingDB),
//             fmt.Sprintf("%.2f", data.AvgNeighborsDB),
//             strconv.Itoa(data.Count5Qi[0]),
//             strconv.Itoa(data.Count5Qi[1]),
//             fmt.Sprintf("%.2f", Bvoice),
//             fmt.Sprintf("%.2f", Bembb/1000), // Convert Kbps to Mbps for display
//             fmt.Sprintf("%.2f", sirAvg),
//             fmt.Sprintf("%.2f", BandwidthUsed),
//             fmt.Sprintf("%.2f", BandwidthUnused),
//         }
//         row = append(row, data.DynamicValues...)
//         row = append(row,
//             strings.Join(data.UeIDsVoice, " "),
//             strings.Join(data.UeIDsEMBB, " "),
//             strings.Join(data.UEThroughputVoice, " "),
//             strings.Join(data.UEThroughputEMBB, " "))
//         array = append(array, row)
//     }

//     CSVFileWriter(array)
// }





/// ********************* Final version working ***************************


// type CellData struct {
//     AvgSIR         float64
//     Count5Qi       [2]int
//     RSRPServingDB  float64
//     AvgNeighborsDB float64
//     DynamicValues  []string
// }


// func callGetRsrpReports(client pb.CCOMonitoringServiceClient, columns map[Column]bool, ocnMap map[Cgi]map[Column]string) {
//     var array [][]string
//     sirMap := make(map[string][]float64)
//     cellData := make(map[string]*CellData)

//     stream, err := client.GetRsrpReports(context.Background(), &pb.NoParam{})
//     if err != nil {
//         log.Printf("Could not get rsrp_info_client: %v", err)
//         return
//     }

//     // Constructing headers dynamically based on sorted column keys
//     dynamicHeaders := make([]string, 0, len(columns))
//     for column := range columns {
//         dynamicHeaders = append(dynamicHeaders, string(column))
//     }
//     sort.Strings(dynamicHeaders) // Sort the column headers
//     headers := append([]string{"UeID", "Serving ID", "rsrpServingDB", "avgNeighborsDB", "count5Qi[1]", "count5Qi[2]", "Bvoice", "Bembb", "Avg SIR", "BandwidthUsed", "BandwidthUnused"}, dynamicHeaders...)
//     array = append(array, headers)

//     for {
//         info, err := stream.Recv()
//         if err == io.EOF {
//             break
//         }
//         if err != nil {
//             log.Printf("Error while streaming %v", err)
//             return
//         }

//         if _, ok := cellData[info.Cgi]; !ok {
//             cellData[info.Cgi] = &CellData{
//                 DynamicValues: make([]string, len(dynamicHeaders)), // Ensure DynamicValues length matches the number of dynamic headers
//             }
//         }
//         data := cellData[info.Cgi]

//         rsrpNeighbors := fmt.Sprint(info.RsrpNeighbors)[4 : len(fmt.Sprint(info.RsrpNeighbors))-1]
//         neighbors := strings.Split(rsrpNeighbors, " ")
//         var linearSum float64
//         for _, neighbor := range neighbors {
//             rsrp := strings.Split(neighbor, ":")[1]
//             n, _ := strconv.Atoi(rsrp)
//             linear := math.Pow(10, float64(n)/10)
//             linearSum += linear
//         }
//         data.AvgNeighborsDB = 10 * math.Log10(linearSum / float64(len(neighbors)))
//         data.RSRPServingDB = float64(info.RsrpServing)
//         sir := data.RSRPServingDB - data.AvgNeighborsDB

//         sirMap[info.Cgi] = append(sirMap[info.Cgi], sir)
//         if info.Fiveqi == 1 {
//             data.Count5Qi[0]++
//         } else if info.Fiveqi == 2 {
//             data.Count5Qi[1]++
//         }

//         // Capture dynamic column values
//         columnMap := ocnMap[Cgi(info.Cgi)]
//         for idx, col := range dynamicHeaders {
//             val, ok := columnMap[Column(col)]
//             if ok {
//                 data.DynamicValues[idx] = val
//             } else {
//                 data.DynamicValues[idx] = "0"
//             }
//         }
//     }

//     // Output results
//     for id, data := range cellData {
//         total := 0.0
//         for _, s := range sirMap[id] {
//             total += s
//         }
//         sirAvg := total / float64(len(sirMap[id]))
//         Bvoice := Cvoice / math.Log2(1+math.Pow(10, sirAvg/10))
//         Bembb := Cembb / math.Log2(1+math.Pow(10, sirAvg/10))

//         BandwidthUsed := float64(data.Count5Qi[0])*Bvoice + float64(data.Count5Qi[1])*Bembb
//         BandwidthUnused := Bcell - BandwidthUsed

//         row := []string{
//             "", // Placeholder for UeID
//             id,
//             fmt.Sprintf("%.2f", data.RSRPServingDB),
//             fmt.Sprintf("%.2f", data.AvgNeighborsDB),
//             strconv.Itoa(data.Count5Qi[0]),
//             strconv.Itoa(data.Count5Qi[1]),
//             fmt.Sprintf("%.2f", Bvoice),
//             fmt.Sprintf("%.2f", Bembb/1000), // Convert Kbps to Mbps for display
//             fmt.Sprintf("%.2f", sirAvg),
//             fmt.Sprintf("%.2f", BandwidthUsed),
//             fmt.Sprintf("%.2f", BandwidthUnused),
//         }
//         row = append(row, data.DynamicValues...)
//         array = append(array, row)
//     }

//     CSVFileWriter(array)
// }



// ************* First grouping if CGI_cellis draft of caculation starts here ****************


// func callGetRsrpReports(client pb.CCOMonitoringServiceClient, columns map[Column]bool, ocnMap map[Cgi]map[Column]string) {
//     var array [][]string
//     sirMap := make(map[string][]float64) // Initialize map to store SIR values for each serving cell ID
//     count5Qi := make(map[string][2]int)  // Initialize map to store counts of 5Qi values 1 and 2 for each serving cell ID

//     stream, err := client.GetRsrpReports(context.Background(), &pb.NoParam{})
//     if err != nil {
//         log.Printf("Could not get rsrp_info_client: %v", err)
//         return
//     }

//     headers := []string{"Time", "UeID", "FiveQi", "S_CGI", "RsrpServing", "RsrpNeigbors", "SIR"}
//     for column := range columns {
//         headers = append(headers, string(column))
//     }
//     array = append(array, headers)
//     for {
//         info, err := stream.Recv()
//         if err == io.EOF {
//             break
//         }
//         if err != nil {
//             log.Printf("Error while streaming %v", err)
//             return
//         }

//         rsrpNeighbors := fmt.Sprint(info.RsrpNeighbors)[4 : len(fmt.Sprint(info.RsrpNeighbors))-1]
//         neighbors := strings.Split(rsrpNeighbors, " ")
//         // var sumNeighbors int
//         var linearSum float64
//         for _, neighbor := range neighbors {
//             rsrp := strings.Split(neighbor, ":")[1]
//             n, _ := strconv.Atoi(rsrp)
//             linear := math.Pow(10, float64(n)/10)
//             linearSum += linear
//         }
//         avgNeighborsLinear := linearSum / float64(len(neighbors))
//         avgNeighborsDB := 10 * math.Log10(avgNeighborsLinear)
//         rsrpServingDB := float64(info.RsrpServing)
//         sir := rsrpServingDB / avgNeighborsDB

//         columnMap := ocnMap[Cgi(info.Cgi)]
//         row := []string{
//             info.Time,
//             info.Ueid,
//             strconv.Itoa(int(info.Fiveqi)),
//             info.Cgi,
//             strconv.Itoa(int(info.RsrpServing)), // Convert int32 to string directly
//             rsrpNeighbors,
//             fmt.Sprintf("%.2f", sir),
//         }
//         for column := range columns {
//             val, ok := columnMap[column]
//             if ok {
//                 row = append(row, val)
//             } else {
//                 row = append(row, "0")
//             }
//         }

//         array = append(array, row)

//         // Update SIR and 5Qi counts for the unique serving cell ID
//         sirMap[info.Cgi] = append(sirMap[info.Cgi], sir)
//         countQi := count5Qi[info.Cgi]
//         if info.Fiveqi == 1 {
//             countQi[0]++  // Increment count of 5Qi value 1
//         } else if info.Fiveqi == 2 {
//             countQi[1]++  // Increment count of 5Qi value 2
//         }
//         count5Qi[info.Cgi] = countQi
//     }

//     // Compute SIRavg for each unique rsrpservingid
//     for id, sirs := range sirMap {
//         total := 0.0
//         for _, s := range sirs {
//             total += s
//         }
//         sirAvg := total / float64(len(sirs))
// 		Bvoice := Cvoice / math.Log2(1+math.Pow(10, sirAvg/10))
//         Bembb := Cembb / math.Log2(1+math.Pow(10, sirAvg/10))
// 		countVoice := count5Qi[id][0]
//         countEmb := count5Qi[id][1]
//         BandwidthUsed := float64(countVoice)*Bvoice + float64(countEmb)*Bembb
// 		BandwidthUnused := Bcell - BandwidthUsed
//         // fmt.Printf("Average SIR for serving ID %s: %.2f\n", id, sirAvg)
//         // fmt.Printf("Count of 5Qi=1 for ID %s: %d\n", id, count5Qi[id][0])
//         // fmt.Printf("Count of 5Qi=2 for ID %s: %d\n", id, count5Qi[id][1])
//         fmt.Printf("Serving ID %s: 5Qi=1 %d: 5Qi=2 %d: Avg SIR: %.2f dB, Bvoice: %.2f Kbps, Bembb: %.2f Mbps, BandwidthUsed: %.2f Kbps, BandwidthUnused: %.2f Kbps\n", id, count5Qi[id][0], count5Qi[id][1], sirAvg, Bvoice, Bembb/1000, BandwidthUsed, BandwidthUnused)

//     }

//     CSVFileWriter(array)
// }


//********* original callGetRsrpReports code with wrong implementation of Ocn values is below ***************

func callGetRsrpReports(client pb.CCOMonitoringServiceClient, columns map[Column]bool, ocnMap map[Cgi]map[Column]string) {
    var array [][]string

    stream, err := client.GetRsrpReports(context.Background(), &pb.NoParam{})
    if err != nil {
        // Log the error and return to allow the caller to handle the situation
        log.Printf("Could not get rsrp_info_client: %v", err)
        return
    }

    headers := []string{"Time", "UeID", "FiveQi", "S_CGI", "RsrpServing", "RsrpNeigbors"}
    for column := range columns {
        headers = append(headers, string(column))
    }
    array = append(array, headers)
    for {
        info, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            // Log the error and return to allow the caller to handle the situation
            log.Printf("Error while streaming %v", err)
            return
        }

        rsrpNeighbors := fmt.Sprint(info.RsrpNeighbors)[4 : len(fmt.Sprint(info.RsrpNeighbors))-1]
        columnMap := ocnMap[Cgi(info.Cgi)]
        row := []string{info.Time, info.Ueid, strconv.Itoa(int(info.Fiveqi)), info.Cgi, strconv.Itoa(int(info.RsrpServing)), rsrpNeighbors}
		fmt.Printf("row %+v\n", row)
		for column := range columns {
            val, ok := columnMap[column]
            if ok {
                row = append(row, val)
            } else {
                row = append(row, "0")
            }
        }

        array = append(array, row)
    }
    CSVFileWriter(array)
}


type Report struct {
	Time          string `bson:"time"`
	Ueid          string `bson:"ueid"`
	FiveQi        string `bson:"fiveQi"`
	Cgi           string `bson:"cgi"`
	RsrpServing   string `bson:"rsrpServing"`
	RsrpNeighbors string `bson:"rsrpNeighbors"`
	CellsPTX      string `bson:"cellsPTX"`
}


func CountFiveQi_new(reports []Report) {
	// Initialize the map to store the count of FiveQi for each cgi
	countMap := make(map[string]map[int]int)
	// Initialize the map to store the count of cgis that meet the threshold condition
	cgiCount := make(map[string]int)
	// Initialize the map to store the set of cgis that meet the 90% occurrence condition
	cgiSet := make(map[string][]string)

	// Iterate over each report in the slice
	for _, report := range reports {
		// Convert fiveQi from string to int
		fiveQi, err := strconv.Atoi(report.FiveQi)
		if err != nil {
			log.Fatalf("Error converting FiveQi to int: %v", err)
		}

		// Initialize the countMap entry for the cgi if it doesn't exist
		if _, ok := countMap[report.Cgi]; !ok {
			countMap[report.Cgi] = make(map[int]int)
		}

		// fmt.Println("count of CGI %s and 5QI %d : %d", report.Cgi,fiveQi, countMap[report.Cgi][fiveQi])
		// fmt.Println("the report is %d", report)

		// Increment the count for the fiveQi for the cgi
		countMap[report.Cgi][fiveQi]++

	}
	// Process the counts to find cgis that meet the conditions
	//fmt.Println("the countmap is %d", countMap)

	fmt.Println("the range of countmap is %d", len(countMap))
	for cgi, counts := range countMap {
		//fmt.Println("counts is %d", counts)
		numUEvoice := counts[1]                                                      // Assuming entry 1 is numUEvoice
		numUEembb := counts[2]                                                       // Assuming entry 2 is numUEembb
		capacity := (100 - (float64(numUEembb)*2 + float64(numUEvoice)*0.105)) / 100 // Adjust calculation as necessary

		fmt.Printf("Capacity for cgi id %s is %f\n", cgi, capacity)
		if capacity < 0.1 {
			cgiCount[cgi]++
		}
	}
	fmt.Println("the cgiCount is %d", cgiCount)
	// Check if each cgi value meets the 90% occurrence condition
	for cgi, count := range cgiCount {
		fmt.Println("the cgi is %d", cgi)
		fmt.Println("the count is %d", count)
		//fmt.Println("the length of reports is %d", len(reports))
		if float64(count) >= 0.9 { // Adjusted to compare against the total reports length
			cgiSet["cgis"] = append(cgiSet["cgis"], cgi)
		}
	}

	fmt.Println("The cgiSet is %v", cgiSet["cgis"])
	// return cgiSet
}

func callGetCellInfo(client pb.CCOMonitoringServiceClient) {

	stream, err := client.GetCellInfo(context.Background(), &pb.NoParam{})
	if err != nil {
		log.Fatalf("Could not get rsrp_info_client: %v", err)
	}

	for {
		info, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while streaming %v", err)
		}
		cgi := info.Cgi
		ptx := info.Ptx
		cellInfo[cgi] = ptx
		fmt.Printf("Cell %v, Power %v \n", info.Cgi, info.Ptx)

	}

}

func callSetCellPTX(client pb.CCOMonitoringServiceClient, cellCGI string, ptx float32) (response string) {

	res, err := client.SetCellPTX(context.Background(), &pb.CellInfo{
		Cgi: cellCGI,
		Ptx: ptx,
	})

	if err != nil {
		log.Fatalf("Could not set cell ptx: %v", err)
	}

	fmt.Println("The response is %v", res.Response)
	if res.Response == "Updated" {
		cellInfo[cellCGI] = ptx
		return res.Response
	} else {
		return "Not Updated"
	}
}

func callGetOcn(client pb.CCOMonitoringServiceClient) map[string]*pb.OcnRecord {
	res, err := client.GetOcn(context.Background(), &pb.GetOcnRequest{})
	if err != nil {
		log.Fatalf("Could not get Ocn: %v", err)
	}
	// Assuming you want to print or use the OcnMap somehow. Here, I'll convert it to a string.
	// Adjust this part based on how you want to use the OcnMap data.
	return res.OcnMap
}

func CSVFileWriter(data [][]string) error {

	var writer *csv.Writer

	// Check if the file exists
	check_file, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		log.Printf("File with the following path %v does not exist \n", filePath)
		log.Print(err)
		return err

	} else {

		// File exists, check if it is empty or not to add headers
		if check_file.Size() <= 1 {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				return err
			}
			defer file.Close()
			// Create a CSV writer
			writer = csv.NewWriter(file)
			//writer.Write(headers)
			writer.WriteAll(data)
			writer.Flush()
			return writer.Error()
		} else {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				return err
			}
			defer file.Close()
			// Create a CSV writer
			writer = csv.NewWriter(file)
			writer.WriteAll(data)
			writer.Flush()
			return writer.Error()
		}
	}
}

// function to format cells ptx data for csv writing
// func getPtxData() string {
// 	var ptxinfo string
// 	for _, cgi := range cgistring {
// 		ptxinfo += cgi + ":" + fmt.Sprint(cellInfo[cgi]) + " "
// 	}
// 	return ptxinfo
// }
// Global MongoDB client
// var client *mongo.Client

// func initMongoDB() {
// 	// MongoDB connection string
//     uri := "mongodb://root:DO4J1wbjte@localhost:27017"
// 	var err error

// 	clientOptions := options.Client().ApplyURI(uri)
// 	client, err = mongo.Connect(context.TODO(), clientOptions)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Check the connection
// 	err = client.Ping(context.TODO(), nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Connected to MongoDB!")
// }

// func saveRsrpReportsToMongo(data [][]string) {
// 	collection := client.Database("RSRPData-3").Collection("rsrpReports-3")

// 	for _, row := range data {
// 		doc := bson.D{
// 			{"time", row[0]},
// 			{"ueid", row[1]},
// 			{"fiveQi", row[2]},
// 			{"cgi", row[3]},
// 			{"rsrpServing", row[4]},
// 			{"rsrpNeighbors", row[5]},
// 			{"cellsPTX", row[6]},
// 		}
// 		_, err := collection.InsertOne(context.TODO(), doc)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// 	fmt.Println("Data inserted into MongoDB")
// }

// func callGetRsrpReports(client pb.CCOMonitoringServiceClient, ocnMap map[string][]OcnRow) {


// func fetchRsrpReportsFromMongo() {

// 	collection := client.Database("RSRPData-3").Collection("rsrpReports-3")

//     ctx := context.TODO()
//     cursor, err := collection.Find(ctx, bson.M{})
//     if err != nil {
//         log.Fatal(err)
//     }
//     defer cursor.Close(ctx)

//     var reports []Report
//     if err = cursor.All(ctx, &reports); err != nil {
//         log.Fatal(err)
//     }

//     CountFiveQi_new(reports)
// }
// func clearMongoDBCollection() {
//     collection := client.Database("RSRPData-3").Collection("rsrpReports-3")
//     ctx := context.TODO()

//     if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
//         log.Fatalf("Error clearing MongoDB collection: %v", err)
//     }
// }

// func getPtxData() string {
// 	var ptxinfo string
// 	for _, cgi := range cgistring {
// 		ptxinfo += cgi + ":" + fmt.Sprint(cellInfo[cgi]) + " "
// 	}
// 	return ptxinfo
// }

// func CountFiveQi_new() (int, int) {
// 	// Initialize counts for 1 and 2
// 	count1 := 0
// 	count2 := 0

// 	// Connect to MongoDB
// 	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:DO4J1wbjte@localhost:27017"))
// 	if err != nil {
// 		log.Fatalf("Could not create MongoDB client: %v", err)
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	err = client.Connect(ctx)
// 	if err != nil {
// 		log.Fatalf("Could not connect to MongoDB: %v", err)
// 	}
// 	defer client.Disconnect(ctx)

// 	// Access the database and collection
// 	database := client.Database("RSRPData")
// 	collection := database.Collection("rsrpReports")

// 	// Query the collection
// 	cursor, err := collection.Find(ctx, bson.M{})
// 	if err != nil {
// 		log.Fatalf("Could not query MongoDB: %v", err)
// 	}
// 	defer cursor.Close(ctx)

// 	// Iterate through the results
// 	for cursor.Next(ctx) {
// 		var result bson.M
// 		err := cursor.Decode(&result)
// 		if err != nil {
// 			log.Fatalf("Could not decode result: %v", err)
// 		}

// 		// Get the value of the 'fiveQi' field
// 		//fmt.Println("The fiveQi is %v",result["fiveQi"])
// 		//fmt.Println("The type of fiveQi is %v",reflect.TypeOf(result["fiveQi"]))
// 		//fiveQi, ok := result["fiveQi"].(string) //(int)
// 		fiveQi := result["fiveQi"].(string)
// 		// if !ok {
// 		// 	log.Fatalf("Invalid type for 'fiveQi' field")
// 		// }

// 		// Increment the count based on the value of 'fiveQi'
// 		if fiveQi == "1" {
// 			count1++
// 		} else if fiveQi == "2" {
// 			count2++
// 		}
// 	}

// 	// Print and log the counts
// 	// fmt.Printf("Count of 1s: %d\n", count1)
// 	// fmt.Printf("Count of 2s: %d\n", count2)
// 	// log.Infof("Count of 1s: %d\n", count1)
// 	// log.Infof("Count of 2s: %d\n", count2)

// 	// Return the counts
// 	return count1, count2
// }

/*func CountFiveQi_new() map[string]map[int]int {

	//ctx := context.TODO()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Access the collection
	collection := client.Database("RSRPData").Collection("rsrpReports")

	// Define a filter
	filter := bson.D{{}}

	// Find the documents in the collection
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}


	// Create a map to store the count of FiveQi for each cgi
	countMap := make(map[string]map[int]int)

	// Iterate over each document in the collection
	//cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Error finding documents in collection: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		err := cursor.Decode(&doc)
		if err != nil {
			log.Fatalf("Error decoding document: %v", err)
		}

		// Extract the cgi and fiveQi values from the document
		cgi := doc["cgi"].(string)
		fiveQiStr := doc["fiveQi"].(string)

		// Convert fiveQiStr to an integer
		fiveQi, err := strconv.Atoi(fiveQiStr)
		if err != nil {
			log.Fatalf("Error converting fiveQi to int: %v", err)
		}

		// Initialize the countMap entry for the cgi if it doesn't exist
		if _, ok := countMap[cgi]; !ok {
			countMap[cgi] = make(map[int]int)
		}

		// Increment the count for the fiveQi for the cgi
		countMap[cgi][fiveQi]++
		fmt.Printf("The count of %d is %d \n", fiveQi, countMap[cgi][fiveQi])
		fmt.Println(countMap)
	}

	return countMap
}*/

// Dada code

// func CountFiveQi_new() map[string][]string {
// 	// Initialize the map to store the count of FiveQi for each cgi
// 	countMap := make(map[string]map[int]int)
// 	// Initialize the map to store the count of cgis that meet the threshold condition
// 	cgiCount := make(map[string]int)
// 	// Initialize the map to store the set of cgis that meet the 90% occurrence condition
// 	cgiSet := make(map[string][]string)
// 	//Connect to MongoDB
// 	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:DO4J1wbjte@localhost:27017"))
// 	if err != nil {
// 		log.Fatalf("Could not create MongoDB client: %v", err)
// 	}
// 	fmt.Println("before 10 seconds")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	fmt.Println("after 10 seconds")
// 	defer cancel()
// 	err = client.Connect(ctx)
// 	if err != nil {
// 		log.Fatalf("Could not connect to MongoDB: %v", err)
// 	}
// 	defer client.Disconnect(ctx)

// 	// Access the collection
// 	collection := client.Database("RSRPData").Collection("rsrpReports")

// 	// Define a filter to get data for the last 10 seconds
// 	// Note: Assuming the "time" field is in Unix timestamp format
// 	filter := bson.D{{"time", bson.D{{"$gt", time.Now().Add(-10 * time.Second)}}}}

// 	// Find the documents in the collection
// 	cursor, err := collection.Find(context.Background(), filter)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer cursor.Close(context.Background())

// 	// Iterate over each document in the collection
// 	for cursor.Next(context.Background()) {
// 		var doc bson.M
// 		err := cursor.Decode(&doc)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		// Extract the cgi and fiveQi values from the document
// 		cgi := doc["cgi"].(string)
// 		fiveQiStr := doc["fiveQi"].(string)

// 		// Convert fiveQiStr to an integer
// 		fiveQi, err := strconv.Atoi(fiveQiStr)
// 		if err != nil {
// 			log.Fatalf("Error converting fiveQi to int: %v", err)
// 		}

// 		// Initialize the countMap entry for the cgi if it doesn't exist
// 		if _, ok := countMap[cgi]; !ok {
// 			countMap[cgi] = make(map[int]int)
// 		}

// 		// Increment the count for the fiveQi for the cgi
// 		countMap[cgi][fiveQi]++
// 	}

// 	// Calculate the Capacity for each cgi
// 	for cgi, count := range countMap {
// 		// Calculate the Capacity for each cgi
// 		// Assuming numUEembb and numUEvoice are available as global variables
// 		numUEvoice := count[1] // Assuming entry 1 is numUEvoice
// 		numUEembb := count[2]  // Assuming entry 2 is numUEembb
// 		capacity := 20 - (float64(numUEembb)*2 + float64(numUEvoice)*0.105) // Assuming BW_embb is 2 MHz and BW_voice is 105 kHz
// 		// Check if the Capacity is less than the threshold
// 		//threshold := 10.0
// 		fmt.Println("Capacity for cgi id %v is %f",cgi, capacity)
// 		if capacity < 10 {
// 			cgiCount[cgi]++
// 		}
// 	}
// 	// Check if each cgi value appears more than 90% of the time
// 	for cgi, count := range cgiCount {
// 		if float64(count) >= 0.9*float64(len(countMap[cgi])) {
// 			cgiSet["cgis"] = append(cgiSet["cgis"], cgi)
// 		}
// 	}

// 	fmt.Println("The cgiSet is %v",cgiSet)
// 	return cgiSet
// }

// func countFiveQi() {
// 	// Initialize MongoDB client
// 	initMongoDB()

// 	// Fetch data from MongoDB
// 	fetchRsrpReportsFromMongo()

// 	// Use the aggregation framework to count the occurrences of each individual value of "fiveQi"
// 	// Aggregate by the "fiveQi" field and count the occurrences of each value
// 	pipeline := bson.D{
// 		{"$group", bson.D{
// 			{"_id", "$fiveQi"},
// 			{"count", bson.D{{"$sum", 1}}},
// 		}},
// 	}

// 	// Execute the aggregation pipeline
// 	collection := client.Database("RSRPData").Collection("rsrpReports")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{pipeline})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer cursor.Close(ctx)

// 	// Print and log the results
// 	for cursor.Next(ctx) {
// 		var result bson.M
// 		err := cursor.Decode(&result)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println(result)
// 		fmt.Println("Getting data from MongoDB Successfully!!")

// 	}
// 	if err := cursor.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// }



// this part gives csv file

// function to format cells ptx data for csv writing

// func dumpDataToMongoDB(data [][]string) {
//     // Connect to MongoDB
//     client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.85.33:27017"))
//     if err != nil {
//         log.Fatalf("Could not create MongoDB client: %v", err)
//     }
// 	log.Infof("Create MongoDB client")
//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()
//     err = client.Connect(ctx)
//     if err != nil {
//         log.Fatalf("Could not connect to MongoDB: %v", err)
//     }
// 	log.Infof("Connected to MongoDB")
//     defer client.Disconnect(ctx)

//     // Access the database and collection
//     database := client.Database("my-release-mongodb")
//     collection := database.Collection("ts-info")
// 	    // Convert data to a BSON document
// 		bsonData := bson.D{}
// 		for _, row := range data {
// 			for _, col := range row {
// 				bsonData = append(bsonData, bson.E{Key: col, Value: col})
// 			}
// 		}
// 		log.Infof("Convert data to a BSON document")
// 		log.Info(bsonData)
// 		// Insert data into MongoDB
// 		_, err = collection.InsertOne(ctx, bsonData)
// 		if err != nil {
// 			log.Fatalf("Could not insert data into MongoDB: %v", err)
// 		}
// 	}

// type RsrpInfo struct {
//     Time        string
//     Ueid        string
//     Fiveqi      uint32
//     Cgi         string
//     RsrpServing int32
//     RsrpNeighbors []int32

// }
// func makeDecision(countMap map[string]float64) bool {
//     // Define a threshold for the average count of FiveQi
//     const threshold = 10

//     // Check if the average count of FiveQi for each cell is above the threshold
//     for _, count := range countMap {
//         if count < threshold {
//             return false
//         }
//     }

//     return true
// }

// func CSVFileWriter(data [][]string) error {

// 	var writer *csv.Writer
// 	headers := []string{"Time", "UeID", "FiveQi", "S_CGI", "RsrpServing", "RsrpNeigbors", "Cells PTX"}

// 	// Check if the file exists
// 	check_file, err := os.Stat(filePath)

// 	if os.IsNotExist(err) {
// 		log.Printf("File with the following path %v does not exist \n", filePath)
// 		log.Print(err)
// 		return err

// 	} else {

// 		// File exists, check if it is empty or not to add headers
// 		if check_file.Size() <= 1 {
// 			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
// 			if err != nil {
// 				return err
// 			}
// 			defer file.Close()
// 			// Create a CSV writer
// 			writer = csv.NewWriter(file)
// 			writer.Write(headers)
// 			writer.WriteAll(data)
// 			writer.Flush()
// 			return writer.Error()
// 		} else {
// 			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
// 			if err != nil {
// 				return err
// 			}
// 			defer file.Close()
// 			// Create a CSV writer
// 			writer = csv.NewWriter(file)
// 			writer.WriteAll(data)
// 			writer.Flush()
// 			return writer.Error()
// 		}
// 	}
// }

// function to format cells ptx data for csv writing

// func dumpDataToMongoDB(data [][]string) {
//     // Connect to MongoDB
//     client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.85.33:27017"))
//     if err != nil {
//         log.Fatalf("Could not create MongoDB client: %v", err)
//     }
// 	log.Infof("Create MongoDB client")
//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()
//     err = client.Connect(ctx)
//     if err != nil {
//         log.Fatalf("Could not connect to MongoDB: %v", err)
//     }
// 	log.Infof("Connected to MongoDB")
//     defer client.Disconnect(ctx)

//     // Access the database and collection
//     database := client.Database("my-release-mongodb")
//     collection := database.Collection("ts-info")
// 	    // Convert data to a BSON document
// 		bsonData := bson.D{}
// 		for _, row := range data {
// 			for _, col := range row {
// 				bsonData = append(bsonData, bson.E{Key: col, Value: col})
// 			}
// 		}
// 		log.Infof("Convert data to a BSON document")
// 		log.Info(bsonData)
// 		// Insert data into MongoDB
// 		_, err = collection.InsertOne(ctx, bsonData)
// 		if err != nil {
// 			log.Fatalf("Could not insert data into MongoDB: %v", err)
// 		}
// 	}

// type RsrpInfo struct {
//     Time        string
//     Ueid        string
//     Fiveqi      uint32
//     Cgi         string
//     RsrpServing int32
//     RsrpNeighbors []int32

// }
// func makeDecision(countMap map[string]float64) bool {
//     // Define a threshold for the average count of FiveQi
//     const threshold = 10

//     // Check if the average count of FiveQi for each cell is above the threshold
//     for _, count := range countMap {
//         if count < threshold {
//             return false
//         }
//     }

//     return true
// }


