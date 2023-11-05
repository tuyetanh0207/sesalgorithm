package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"strconv"
	"math/rand"
	"time"
	"log"
)

type Config struct {
	Ports []int `json:"ports"`
}

type Timestamp struct {
	Time int
}

type Object struct {
	Timestamp Timestamp
	ProcessID int
}

type ReceivedItem struct {
	ProcessID int
	Timestamps []int
	Content string
}
type BufferItem struct {
	ReceivedList []ReceivedItem
	Timestamp []int
	ProcessID int
}

type Message struct {
	Timestamps    []int
	ReceivedArray []ReceivedItem
	Sender        int
	Content string
}

func loadConfig(filename string) (Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		if err := decoder.Decode(&config); err == io.EOF {
			break
		} else if err != nil {
			return config, err
		}
	}

	return config, nil
}

func sendMessage(senderID, receiverID int, timestamps []int, receivedArray []ReceivedItem, chs []chan Message, messageName string) {
	//fmt.printf("In sendMessage function: prepare to implement sendMessage from process %d to process %d, message is: %s, timestamp is %v\n", senderID, receiverID, messageName, timestamps)
	message := Message{
		Timestamps:    timestamps,
		ReceivedArray: receivedArray,
		Sender:        senderID,
		Content: messageName,
	}
	//fmt.println("In sendMessage function: create message successfully, message is:", message)
	chs[receiverID] <- message
	//fmt.println("In sendMessage function: sent to channel of receiver successfully, message is:", message)
}
// if timestamp1 >= timestamp2 returns true, else false
func compareTwoTimestamps(timestamp1 []int, timestamp2 []int) bool {
	//fmt.println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2)

	for i := 0; i < len(timestamp1); i++ {
            if timestamp1[i] < timestamp2[i] {
				//fmt.println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " result is timestamp1 >= timestamp2, return false")

                return false
            }
    }
	//fmt.println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " result is timestamp1 < timestamp2, return true")

	return true
}
func mergeTwoTimestamps(timestamp1 []int, timestamp2 []int, receiverID int) []int {
	//fmt.println("In mergeTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID)
	//tempTimestamp1 := timestamp1
	timestamp1[receiverID]++
	//fmt.println("In mergeTwoTimestamps function: timestamp1 after incremented is:", timestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID)
	for i := 0; i < len(timestamp1); i++ {
        if timestamp2[i] > timestamp1[i] {
			timestamp1[i] = timestamp2[i]
		}
	}
	//fmt.println("In mergeTwoTimestamps function: timestamp1 is:", tempTimestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID, " final timestamp after merge is:", timestamp1)

	////fmt.println("In mergeTwoTimestamps function: final timestamp after merge is:", timestamp1)

	return timestamp1
}
func checkSentBefore(carriers []ReceivedItem, receiverID int) bool {
	//fmt.println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID)

	for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == receiverID {
			//fmt.println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID == receiverID, return true")	
            return true
        }
    }
	//fmt.println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID!= receiverID, return false")	
    return false

}
func sentBeforeAt(carriers []ReceivedItem, receiverID int) int {
	//fmt.println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID)

    for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == receiverID {
            //fmt.println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID == receiverID, return i")    
            return i
        }
    }
    //fmt.println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID!= receiverID, return -1")    
    return -1
}

func mergeTwoCarriers(senderCarriers []ReceivedItem, receiverCarriers []ReceivedItem) []ReceivedItem {
	println("need to merge two carriers")
	//fmt.println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " receiverCarriers is:", receiverCarriers)
	 for i := 0; i < len(senderCarriers); i++ {
        for j := 0; j < len(receiverCarriers); j++ {
            if senderCarriers[i].ProcessID != receiverCarriers[j].ProcessID {
				//fmt.println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " prepare to append receiverCarriers is:", receiverCarriers)
                senderCarriers = append(senderCarriers, receiverCarriers[j])
            }
			// else {
			// 	// TODO:
			// 	//fmt.println("..")
			// }
        }
    }
	//fmt.println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " receiverCarriers is:", receiverCarriers, "final senderCarriers is:", senderCarriers)
	
	////fmt.println("In mergeTwoCarriers function: final senderCarriers is:", senderCarriers)
    return senderCarriers
}
func getTimestampOfProcessInCarrier(processID int, carriers []ReceivedItem) []int{
	//fmt.println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers)

	for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == processID {
			//fmt.println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers, " result is carriers[i].ProcessID == processID, return carriers[i].Timestamps", carriers[i].Timestamps)
            return carriers[i].Timestamps
        }
    }
	//fmt.println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers, " result is carriers[i].ProcessID!= processID, return nil")
    return nil

}
func clearCarriers(carriers []ReceivedItem, processID int, timestamps []int) []ReceivedItem {
	//fmt.println("In clearCarriers function: carriers is:", carriers, " processID is:", processID, " timestamps is:", timestamps)

    for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == processID {
            //fmt.println("In clearCarriers function: carriers is:", carriers, " processID is:", processID, " result is carriers[i].ProcessID == processID, return carriers[i].Timestamps")
            carriers[i].Timestamps = timestamps
        }
    }
    //fmt.println("In clearCarriers function: final carriers is:", carriers)
    return carriers
	
}
func minOfBuffer(buffers []BufferItem, processID int) (BufferItem, ReceivedItem, int, int) {
	//fmt.println("In minOfBuffer function: buffers is:", buffers, " processID is:", processID)
	count:=0
	if len(buffers) >= 1 {
		minBuffer := buffers[0]
		minReceivedItem := buffers[0].ReceivedList[0]
		minTimestamp := buffers[0].ReceivedList[0].Timestamps
		outIdx:=0
		inIdx:=0
		for i:=0; i<len(buffers); i++ {
			for j:=0; j<len(buffers[i].ReceivedList); j++ {
				if buffers[i].ReceivedList[j].ProcessID==processID  {
					if count == 0 {
                    minBuffer = buffers[i]
                    minReceivedItem = buffers[i].ReceivedList[j]
                    minTimestamp = buffers[i].ReceivedList[j].Timestamps
                    outIdx=i
                    inIdx=j
					count++
					continue
					} else {
						if compareTwoTimestamps(buffers[i].ReceivedList[j].Timestamps, minTimestamp){
							minBuffer = buffers[i]
                            minReceivedItem = buffers[i].ReceivedList[j]
                            minTimestamp = buffers[i].ReceivedList[j].Timestamps
                            outIdx=i
                            inIdx=j
						}
					}
                }
			}
		}
		//fmt.println("In minOfBuffer function: find out min of buffer, buffer is:", buffers, " processID is:", processID, " minBuffer is: ", minBuffer, " minReceivedItem is:", minReceivedItem, " outIdx is:", outIdx, " inIdx is: ", inIdx)
		return minBuffer, minReceivedItem, outIdx, inIdx
	}
	//fmt.println("In minOfBuffer function: buffers is:", buffers, " processID is:", processID, "cannot find min buffer, it mean in buffer don't contain processID")

	return BufferItem{}, ReceivedItem{}, -1, -1

}
func removeElement(slice []BufferItem, index int) []BufferItem {
    return append(slice[:index], slice[index+1:]...)
}
// return - 1 when when can find resonable min buffer
func processBuffers(buffers [][]BufferItem, processID int, timestamps [][]int, carriers [][]ReceivedItem ) (int, string) {
	//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers)
	minBuffer, minReceivedItem, outIdx, inIdx := minOfBuffer(buffers[processID], processID)
	if outIdx < 0  || inIdx < 0 {
		return -1, "Invalid min buffer index"
	}
	//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx)
	
	////fmt.println("In processBuffers function: minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx)
	message := Message{
		Timestamps:    minReceivedItem.Timestamps,
        ReceivedArray: minBuffer.ReceivedList,
        Sender:        minBuffer.ProcessID,
        Content: minReceivedItem.Content,
	}
	//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " prepare implement receiveMessageProcess in it")
	res := receiveMessageProcessing(timestamps, buffers, processID, minReceivedItem.ProcessID, carriers, message)
	//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess in it, res is ", res)
	
	// delete buffer sent
	////fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess in it, res is ", res)
	
	buffers[processID]=removeElement(buffers[processID], outIdx)
	//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " buffer after cleared is", buffers[processID])
	
	if res ==1 {
		//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess sucessfully in processBuffer ", res)
	
		////fmt.println("In processBuffers function: res of receiveMessageProcessing is", res)
		return 1, "Process buffer of process " + strconv.Itoa(processID) + "from process " +strconv.Itoa(minBuffer.ProcessID)+ " .Message name is: " + minReceivedItem.Content +" successfully."
	} else {
		//fmt.println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess but unsucessfully, in processBuffer ", res)
	
		////fmt.println("In processBuffers function: res of receiveMessageProcessing is", res)
		return 0, "Error in receiving message in process buffer  of process " + strconv.Itoa(processID) + "from process " +strconv.Itoa(minBuffer.ProcessID)+ " .Message name is: " + minReceivedItem.Content +" ."
	}
}
func writeToLogFile(fileNumber int, message string) error {
	// Tạo tên tệp dựa trên số thứ tự
	fileName := fmt.Sprintf("logfile%d.txt", fileNumber)

	// Mở hoặc tạo tệp nhật ký (sử dụng flag os.O_APPEND để ghi tiếp vào tệp đã tồn tại)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ghi thông điệp vào tệp nhật ký
	_, err = fmt.Fprintf(file, "%s\n", message)
	if err != nil {
		return err
	}

	return nil
}
func receiveMessageProcessing(timestamps [][]int, buffers [][]BufferItem, currReceiverID int, senderID int, carriers [][]ReceivedItem, message Message) int {
	////fmt.println("Carrier has not been sent before, can accepted message.")
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message)
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 timestamps, timestamp1 is", timestamps[currReceiverID], " timestamp2 is", message.Timestamps, " currReceiverID is", currReceiverID)
	//tempTimestamp := timestamps[currReceiverID]
	////fmt.println("In ReceiveMessageProcessing function: prepare to merge 2 timestamps")
	timestamps[currReceiverID] = mergeTwoTimestamps(timestamps[currReceiverID], message.Timestamps, currReceiverID)
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 timestamps, timestamp of receiver is", tempTimestamp, " timestamp2 is", message.Timestamps, " currReceiverID is", currReceiverID, " timestamp after merge is", timestamps[currReceiverID])
	
	////fmt.println("Timestamps merged:", timestamps[currReceiverID])
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 carriers, carrier 1 is", carriers[currReceiverID], " carrier2 is", message.ReceivedArray)
	
	////fmt.println("In ReceiveMessageProcessing function: prepare to merge 2 carriers")
	carriers[currReceiverID]= mergeTwoCarriers(carriers[currReceiverID], message.ReceivedArray)
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 carriers, carrier 1 is", carriers[currReceiverID], " carrier2 is", message.ReceivedArray, ". carrier after merge is", carriers[currReceiverID])
	////fmt.println("In ReceiveMessageProcessing function: timestamp merged is:", carriers[currReceiverID])
	////fmt.println("In ReceiveMessageProcessing function: prepare to clear carriers")
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to clear carrier, carrier is", carriers[currReceiverID], " senderID is", senderID, " timestamp is: ", message.Timestamps)
	
	carriers[currReceiverID] = clearCarriers(carriers[currReceiverID], senderID, message.Timestamps)
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to clear carrier, carrier is", carriers[currReceiverID], " senderID is", senderID, " timestamp is: ", message.Timestamps, " carrier after clear is: ", carriers[currReceiverID])
	////fmt.println("In ReceiveMessageProcessing function: carriers after clear is:",carriers[currReceiverID])
	////fmt.println("In ReceiveMessageProcessing function: prepare to process buffer")
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to process buffer, buffer is", buffers[currReceiverID], "timestamps is", timestamps, "carcarriers is: ", carriers[currReceiverID])
	
	resProcessBuffer, str :=processBuffers(buffers, currReceiverID, timestamps, carriers)
	//fmt.println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to process buffer, buffer is", buffers[currReceiverID], "timestamps is", timestamps, "carcarriers is: ", carriers[currReceiverID], " resBufferProcess is", resProcessBuffer, str)
	
	////fmt.println("In ReceiveMessageProcessing function: resProcessBuffer is:", resProcessBuffer, str)
	fmt.Println(str)
	if resProcessBuffer==1 {
		//fmt.printf("In ReceiveMessageProcessing function: Process %d received message from Process %d: successfully! Content is: %s., timestamp is %v .\n", currReceiverID, senderID, message.Content, message.Timestamps)
		return 1
	} else {
		return 0
	}
}

type CustomLogger struct {
	consoleLogger *log.Logger
	fileLogger    *log.Logger
}

func NewCustomLogger(logFileName string) (*CustomLogger, error) {
	// Open or create the log file
	file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Create a console logger that writes to standard output
	consoleLogger := log.New(os.Stdout, "", log.LstdFlags)

	// Create a file logger that writes to the opened log file
	fileLogger := log.New(file, "", log.LstdFlags)

	return &CustomLogger{
		consoleLogger: consoleLogger,
		fileLogger:    fileLogger,
	}, nil
}

func (c *CustomLogger) PrintLog(index int, message string) {
	// Log to console
	c.consoleLogger.Println(message)

	// Log to file based on index
	if index >= 1 && index <= len(loggers) {
		loggers[index-1].fileLogger.Println(message)
	} else {
		// Handle index out of range error or log to a default file
		c.consoleLogger.Println("Invalid log index:", index)
	}
}
var loggers []*CustomLogger

func main() {
	//fmt.println("Start program.")
	
	
	//fmt.println("Log files created successfully in 'logfiles' directory.")
	config, err := loadConfig("config.json")
	if err != nil {
		//fmt.println("Error reading config file:", err)
		return
	}

	//fmt.println("Ports:", config.Ports)
	err1 := os.Mkdir("logfiles", os.ModePerm)
	if err1 != nil {
		//fmt.printf("Cannot create directory: %v\n", err)
		return
	}
	err1 = os.Chdir("logfiles")
	if err1 != nil {
		//fmt.printf("Cannot change directory: %v\n", err)
		return
	}
	//var loggers []*CustomLogger
	for i := 0; i < len(config.Ports); i++ {
		logFileName := fmt.Sprintf("logfile%d.txt", i)
		customLogger, err := NewCustomLogger(logFileName)
		if err != nil {
			//fmt.printf("Error creating custom logger for file %s: %v\n", logFileName, err)
			return
		}
		loggers = append(loggers, customLogger)
		defer func(i int) {
			// if e; err != nil {
			// 	//fmt.printf("Error closing log file %s: %v\n", logFileName, err)
			// }
		}(i)
	}

	// Write messages to specific log files based on index
	for i := 0; i < len(config.Ports); i++ {
		loggers[i].PrintLog(i, fmt.Sprintf("Log message for port %d", i+8000-1))
	}


	var wg sync.WaitGroup

	// Initialize channels and timestamps for processes
	channels := make([]chan Message, len(config.Ports))
	buffers := make([][]BufferItem, len(config.Ports))
	carriers := make([][]ReceivedItem, len(config.Ports))
	timestamps := make([][]int, len(config.Ports))
	for i := 0; i < len(config.Ports); i++ {
		channels[i] = make(chan Message)
		buffers[i] = make([]BufferItem, 0)
		carriers[i] = make([]ReceivedItem, 0)
		timestamps[i] = make([]int, len(config.Ports))
		for j := range timestamps[i] {
			timestamps[i][j] = 0
		}

		////fmt.println("carriers new:", carriers[i])
		////fmt.println("timestamps: %d ", i, timestamps)
	}

	// Start processes based on ports from config
	for _, port := range config.Ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			for i := 0; i < len(config.Ports); i++ {
				if i != port-8000 {
					go func(senderID, receiverID int) {
						for {
							message := <-channels[receiverID]
							// Process received message logic
							tempSenderID:= message.Sender - 8000
							//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. prepare process.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray)
							loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. prepare process.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
							// check if carrier of sending processed has been sent before
							//loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf())
							if (!checkSentBefore(message.ReceivedArray, receiverID)) {
								//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray)
								loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
								ress:= receiveMessageProcessing(timestamps, buffers, receiverID, tempSenderID,carriers, message)
								if (ress==1){
									loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., receive message successfully . End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
									//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., receive message successfully . End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray) 
								} else {
									loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., but receive message unsuccessfully.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
									//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., but receive message unsuccessfully.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray) 
								}

							} else{
							// compare 2 timestamps
							////fmt.println("checksentbefore: ",checkSentBefore(message.ReceivedArray, receiverID) )
							timestampsInCarrier := getTimestampOfProcessInCarrier(receiverID, message.ReceivedArray)
							//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentBefore, prepare to process timestamp, timestampIncarrier is: %v.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ) //
							loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentBefore, prepare to process timestamp, timestampIncarrier is: %v.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier))
							if (timestampsInCarrier != nil) {
								
								if compareTwoTimestamps(timestamps[receiverID], timestampsInCarrier) {
									
									res:= receiveMessageProcessing(timestamps, buffers, receiverID, tempSenderID,carriers, message)
									if (res==1){
										loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, receive message successfully %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier))
										//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, receive message successfully %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier) 
									} else {
										loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, but receive message unsuccessfully %v.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ))
										//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, but receive message unsuccessfully %v.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ) 
									}
							
								} else { // add received item to buffer
									// buffer if timestamp is not correct
									
									buffers[receiverID] = append(buffers[receiverID], BufferItem{Timestamp: message.Timestamps, ReceivedList:message.ReceivedArray})
									////fmt.println("Buffer:", buffers[receiverID])
									loggers[tempSenderID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is %v unacceptable, buffer message, current buffer is: %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier, buffers[receiverID] ))
									//fmt.printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is %v unacceptable, buffer message, current buffer is: %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier, buffers[receiverID] ) 
								}
							}
							
							}
							
						}
						//fmt.println("Complete sending section.")
					}(i, port-8000)
				}
			}

			
			for i := 0; i < len(config.Ports); i++ {
				for j := 0; j < 150; j++ {
					receiverID := (port + j -1) % len(config.Ports)
					senderID := port - 8000
					
					
					if (receiverID != senderID) {
						////fmt.println("In sending section:")
						// Chỉ cập nhật timestamps tại vị trí sender tương ứng
						tempTimestamp:=timestamps[senderID]
						//tempTimestamp[senderID]= tempTimestamp[senderID] + 1
						newTimestamp := make([]int, len(tempTimestamp))
						newCarrier := make([]ReceivedItem, len(carriers[senderID]))
						copy(newTimestamp, tempTimestamp)
						copy(newCarrier, carriers[senderID])
						newTimestamp[senderID] = newTimestamp[senderID] + 1
						randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))			
						messageName := "Message " + strconv.Itoa(senderID*len(config.Ports)+receiverID + i + j + randGenerator.Intn(1000) + randGenerator.Intn(1000))
						////fmt.printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v \n", senderID, receiverID, newTimestamp, messageName, carriers[senderID])
						loggers[senderID].PrintLog(senderID, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v \n", senderID, receiverID, newTimestamp, messageName, carriers[senderID]))
						timestamps[senderID][senderID]=newTimestamp[senderID]
						sendMessage(port, receiverID, newTimestamp, newCarrier, channels, messageName)
						
						newItem := ReceivedItem{
							ProcessID:  receiverID,
							Timestamps: newTimestamp,
							Content:    messageName,
						}
						idx := sentBeforeAt(carriers[senderID], receiverID) 
						if idx != -1 {
							////fmt.println("idx of carrier in if block: ", idx)
							newCarrierItem := ReceivedItem{
								ProcessID:  receiverID,
								Timestamps: newTimestamp,
								Content: carriers[senderID][idx].Content,
							}
							carriers[senderID][idx] = newCarrierItem
							////fmt.printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just replace carrier %v at index %d, current carrier is: %v \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newCarrierItem, idx, carriers[senderID])
							loggers[i].PrintLog(i, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just replace carrier %v at index %d, current carrier is: %v \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newCarrierItem, idx, carriers[senderID]))
							//loggers[i].PrintLog(i, fmt.Sprintf())
						
						} else {
							carriers[senderID] = append(carriers[senderID], newItem)
							////fmt.printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just append carrier %v at index %d, current carrier is: %v  \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newItem, idx, carriers[senderID])
							//
							loggers[i].PrintLog(i, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just append carrier %v at index %d, current carrier is: %v  \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newItem, idx, carriers[senderID]))
						}
					
					}
					
				}
			}
			//fmt.println("Complete sending section.")
		}(port)
	}

	wg.Wait()
}
