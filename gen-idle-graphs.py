import csv
import os
import pandas as pd
import re

ALGORITHM = ""
HOMO = ""
LATENCY = ""

# Dictonary of Algorithm dictonaries, which hold node information
CSVDATA = {
}


# Accepts a file path for a folder containing experiment .csvs of any latency for a specific set of homogenous 
# or homogenious experiments of one tested algorithm. Outputs a .csv graph with the idle time data for all the input 
# experiments nodes. 

def main(): 

    # Give code access to folder with files
    folderPathName = input("Input Folder Path: ")

    # Read csvs from file and create dictonary
    read_csv_from_folder(folderPathName)

    format_csvdata(CSVDATA)

    print(CSVDATA)

    # Create and fomat table
    create_table(CSVDATA)




def read_csv_from_folder(folderPathName):

    # walk the entire folder and extract the header, then data
    for root, _, files in os.walk(folderPathName): 
        for file in files: 
            global ALGORITHM
            global LATENCY
            global HOMO
            
            path = os.path.join(root, file)
            f = open(path, "r")
            
            # Remove csv header
            header = f.readline()

            # Use regex to find the algorithm and rate in the path name
            algorithmPattern = r'(?:[^-]+-){3}([^-]+)'
            ratePattern = r'(?:[^-]+-){13}([^-]+)'
            latencyPattern = r'(?:[^-]+-){5}([^-]+)'
            homoPattern = r'(?:[^-]+-){7}([^-]+)'
            fileSizePattern = r'(?:[^-]+-){11}([^-]+)'

            thirdValue = re.search(algorithmPattern, path)
            fithValue = re.search(latencyPattern, path)
            seventhValue = re.search(homoPattern, path)
            thirteenthValue = re.search(ratePattern, path)
            eleventhValue = re.search(fileSizePattern, path)


            ALGORITHM = str(thirdValue.group(1))
            LATENCY = str(fithValue.group(1))
            homo = str(seventhValue.group(1))
            fileSize = str(eleventhValue.group(1))

            #Check if homogenous or heterogenous
            if homo == 'false':
                HOMO = "Heterogeneous"
            else:
                HOMO = "Homogenous"


            # Get rid of trailing .csv
            rate = re.sub(r".{4}$", "", str(thirteenthValue.group(1)))
            print(rate)

            # Seperate csv file by , and save information to dictonary
            reader = csv.reader(f, delimiter=",")
          
            for line in reader: 

                # Make sure to skip 

                # Get data from line
                nodeID = line[0]
                eventType = line[2]
                timeStamp = line[4]
                duration = line[5]
                peer = line[3]

                save_to_dictonary(nodeID, eventType, timeStamp, duration, peer, rate)


def format_csvdata(CSVDATA): 
        
        for experiment in CSVDATA: 
                for node in CSVDATA[experiment]:


                    endTime = int(CSVDATA[experiment][node]["EndDuration"]) + int(CSVDATA[experiment][node]["EndTimestamp"])

                    totalTime = int(endTime) - int(CSVDATA[experiment][node]["StartTimestamp"])

                    idletime = int(totalTime) - int(CSVDATA[experiment][node]["SumDuration"])

                    CSVDATA[experiment][node].clear()

                    CSVDATA[experiment][node] = str(CSVDATA[experiment][node])

                    CSVDATA[experiment][node] = idletime



def save_to_dictonary(nodeID, eventType, timeStamp, duration, peer, rate):

    # Initialize algorithm dict if not already initialized
    if ("Request Rate: "+str(rate)) not in CSVDATA: 
        CSVDATA["Request Rate: "+str(rate)] = {}

    # Check if nested key was created for node, if it isn't, create new dictonary for node, else add to existing node dictonary
    if ("Node " + str(nodeID)) not in CSVDATA["Request Rate: "+str(rate)]: 
        CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)] = {"NodeID": nodeID, "StartTimestamp": 0, "EndTimestamp": timeStamp, "EndDuration": duration, "SumDuration": duration, "Peer": peer, "TransferCount": 0}

    elif ("Node " + str(nodeID)) in CSVDATA["Request Rate: "+str(rate)]:
        
        if (eventType != 'node-offline') and (eventType != 'node-online') : 
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["EndTimestamp"] = timeStamp
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["EndDuration"] = duration
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["SumDuration"] = int(CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["SumDuration"]) + int(duration)

        if(eventType == 'file-transfer'):
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["TransferCount"] = int(CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["TransferCount"]) + 1
    
        # Find first timestamp when peer is the load balencer, make that the start timestamp
        if (peer != '') & (CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["Peer"] == ''): 
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["Peer"] = peer
            CSVDATA["Request Rate: "+str(rate)]["Node " + str(nodeID)]["StartTimestamp"] = timeStamp


def create_table(CSVDATA): 

    # Create the dataframe for the .csv
    df = pd.DataFrame.from_dict(CSVDATA, orient="index")

    # Table title
    df.index.name = "IDLE TIME DATA:  Algorithm: " + ALGORITHM + ", Latency: " + LATENCY + ", " + HOMO
    df.reset_index(inplace=True)

    sortColumns = ["IDLE TIME DATA:  Algorithm: " + ALGORITHM + ", Latency: " + LATENCY + ", " + HOMO] + sorted(df.columns[1:], key=lambda x: int(x.split()[1]))
    df = df[sortColumns]
    
    # Export .csv
    fileName = "Output.xlsx"
    df.to_excel(fileName, sheet_name="Tables", engine='openpyxl')

    print(f"Excel file '{fileName}' created successfully!")

if __name__ == "__main__":
    main()