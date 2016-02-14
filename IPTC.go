package EXIF

import (
	"bytes"
	"fmt"
	"io"
)

/******************************************************************************
*
* Function:     get_IPTC
*
* Description:  Extracts IPTC-NAA IIM data from the string provided, and returns
*               the information as an array
*
* Parameters:   Data_Str - the string containing the IPTC-NAA IIM records. Must
*                          be exact length of the IPTC-NAA IIM data.
*
* Returns:      OutputArray - Array of IPTC-NAA IIM records
*               FALSE - If an error occured in decoding
*
******************************************************************************/
type iptcRecord struct {
	recType          string
	recRecordNumber  byte
	recDataSetNumber byte
	recData          []byte
}

func getIPTC(reader io.Reader) []iptcRecord {

	// Create the array to receive the data
	outputArray := make([]iptcRecord)

	header := make([]byte, 5)

	// Cycle through the IPTC records, decoding and storing them
	for true { // pos < reader.

		// @TODO - Extended Dataset record not supported

		n, err := reader.Read(header)
		if err != nill {
			// Not enough data left for a record - Probably corrupt data - ERROR
			// Change: changed to return partial data as of revision 1.01
			return outputArray
		}

		// Unpack data from IPTC record:
		// First byte - IPTC Tag Marker - always 28
		// Second byte - IPTC Record Number
		// Third byte - IPTC Dataset Number
		// Fourth and fifth bytes - two byte size value
		// $iptc_raw = unpack( "CIPTC_Tag_Marker/CIPTC_Record_No/CIPTC_Dataset_No/nIPTC_Size", substr($Data_Str,$pos) );
		iptcTagMarker := header[0]
		iptcRecordNumber := header[1]
		iptcDataSetNumber := header[2]
		iptcSize := uint16(header[3]) | (uint16(header[4]) << 8)

		// Construct the IPTC type string eg 2:105
		iptcType := fmt.Sprintf("%01d:%02d", iptcRecordNumber, iptcDataSetNumber)

		// Check if there is sufficient data for reading the record contents
		content := make([]byte, iptcSize)
		n, err = reader.Read(content)
		if err != nill {
			// Not enough data left for the record content - Probably corrupt data - ERROR
			// Change: changed to return partial data as of revision 1.01
			return outputArray
		}

		// Add the IPTC record to the output array
		//$OutputArray[] = array( "IPTC_Type" => $iptctype ,
		//                        "RecName" => $GLOBALS[ "IPTC_Entry_Names" ][ $iptctype ],
		//                        "RecDesc" => $GLOBALS[ "IPTC_Entry_Descriptions" ][ $iptctype ],
		//                        "RecData" => substr( $Data_Str, $pos, $iptc_raw['IPTC_Size'] ) );
		record := &iptcRecord{recType: iptcType, recRecordNumber: iptcRecordNumber, recDataSetNumber: iptcDataSetNumber, recData: content}
		outputArray = append(outputArray, record)
	}
	return outputArray
}

/******************************************************************************
* End of Function:     get_IPTC
******************************************************************************/

/******************************************************************************
*
* Function:     put_IPTC
*
* Description:  Encodes an array of IPTC-NAA records into a string encoded
*               as IPTC-NAA IIM. (The reverse of get_IPTC)
*
* Parameters:   new_IPTC_block - the IPTC-NAA array to be encoded. Should be
*                                the same format as that received from get_IPTC
*
* Returns:      iptc_packed_data - IPTC-NAA IIM encoded string
*
******************************************************************************/

func putIPTC(iptcRecords []iptcRecord) ([]byte, bool) {
	// Initialise the output
	var iptcData bytes.Buffer

	// Cycle through each record in the new IPTC block
	for _, record := range iptcRecords {

		// Write the IPTC-NAA IIM Tag Marker, Record Number, Dataset Number and Data Size to the packed output data string
		iptcData.Write(byte(28))
		iptcData.Write(record.recRecordNumber)
		iptcData.Write(record.recDataSetNumber)
		iptcData.Write(uint16(len(record.recData)))
		iptcData.Write(record.recData)
		//$iptc_packed_data .= pack( "CCCn", 28, $IPTC_Record, $IPTC_Dataset, strlen($record['RecData']) );

		// Write the IPTC-NAA IIM Data to the packed output data string
		//$iptc_packed_data .= $record['RecData'];

	}
	// Return the IPTC-NAA IIM data
	return iptcData.Bytes(), true
}

/******************************************************************************
* End of Function:     put_IPTC
******************************************************************************/

/******************************************************************************
* Global Variable:      IPTC_Entry_Names
*
* Contents:     The names of the IPTC-NAA IIM fields
*
******************************************************************************/

var aIPTCEntryNames = map[uint16]string{
	1*256 + 0:   "Model Version",
	1*256 + 5:   "Destination",
	1*256 + 20:  "File Format",
	1*256 + 22:  "File Format Version",
	1*256 + 30:  "Service Identifier",
	1*256 + 40:  "Envelope Number",
	1*256 + 50:  "Product ID",
	1*256 + 60:  "Envelope Priority",
	1*256 + 70:  "Date Sent",
	1*256 + 80:  "Time Sent",
	1*256 + 90:  "Coded Character Set",
	1*256 + 100: "UNO (Unique Name of Object)",
	1*256 + 120: "ARM Identifier",
	1*256 + 122: "ARM Version",

	// Application Record
	2*256 + 0:   "Record Version",
	2*256 + 3:   "Object Type Reference",
	2*256 + 5:   "Object Name (Title)",
	2*256 + 7:   "Edit Status",
	2*256 + 8:   "Editorial Update",
	2*256 + 10:  "Urgency",
	2*256 + 12:  "Subject Reference",
	2*256 + 15:  "Category",
	2*256 + 20:  "Supplemental Category",
	2*256 + 22:  "Fixture Identifier",
	2*256 + 25:  "Keywords",
	2*256 + 26:  "Content Location Code",
	2*256 + 27:  "Content Location Name",
	2*256 + 30:  "Release Date",
	2*256 + 35:  "Release Time",
	2*256 + 37:  "Expiration Date",
	2*256 + 35:  "Expiration Time",
	2*256 + 40:  "Special Instructions",
	2*256 + 42:  "Action Advised",
	2*256 + 45:  "Reference Service",
	2*256 + 47:  "Reference Date",
	2*256 + 50:  "Reference Number",
	2*256 + 55:  "Date Created",
	2*256 + 60:  "Time Created",
	2*256 + 62:  "Digital Creation Date",
	2*256 + 63:  "Digital Creation Time",
	2*256 + 65:  "Originating Program",
	2*256 + 70:  "Program Version",
	2*256 + 75:  "Object Cycle",
	2*256 + 80:  "By-Line (Author)",
	2*256 + 85:  "By-Line Title (Author Position) [Not used in Photoshop 7]",
	2*256 + 90:  "City",
	2*256 + 92:  "Sub-Location",
	2*256 + 95:  "Province/State",
	2*256 + 100: "Country/Primary Location Code",
	2*256 + 101: "Country/Primary Location Name",
	2*256 + 103: "Original Transmission Reference",
	2*256 + 105: "Headline",
	2*256 + 110: "Credit",
	2*256 + 115: "Source",
	2*256 + 116: "Copyright Notice",
	2*256 + 118: "Contact",
	2*256 + 120: "Caption/Abstract",
	2*256 + 122: "Caption Writer/Editor",
	2*256 + 125: "Rasterized Caption",
	2*256 + 130: "Image Type",
	2*256 + 131: "Image Orientation",
	2*256 + 135: "Language Identifier",
	2*256 + 150: "Audio Type",
	2*256 + 151: "Audio Sampling Rate",
	2*256 + 152: "Audio Sampling Resolution",
	2*256 + 153: "Audio Duration",
	2*256 + 154: "Audio Outcue",
	2*256 + 200: "ObjectData Preview File Format",
	2*256 + 201: "ObjectData Preview File Format Version",
	2*256 + 202: "ObjectData Preview Data",

	// Pre-ObjectData Descriptor Record
	7*256 + 10: "Size Mode",
	7*256 + 20: "Max Subfile Size",
	7*256 + 90: "ObjectData Size Announced",
	7*256 + 95: "Maximum ObjectData Size",

	// ObjectData Record
	8*256 + 10: "Subfile",

	// Post ObjectData Descriptor Record
	9*256 + 10: "Confirmed ObjectData Size",
}

/******************************************************************************
* End of Global Variable:     IPTC_Entry_Names
******************************************************************************/

/******************************************************************************
* Global Variable:      IPTC_Entry_Descriptions
*
* Contents:     The Descriptions of the IPTC-NAA IIM fields
*
******************************************************************************/

var aIPTCEntryDescriptions = map[uint16]string{
	// Envelope Record
	1*256 + 0:   "2 byte binary version number",
	1*256 + 5:   "Max 1024 characters of Destination",
	1*256 + 20:  "2 byte binary file format number, see IPTC-NAA V4 Appendix A",
	1*256 + 22:  "Binary version number of file format",
	1*256 + 30:  "Max 10 characters of Service Identifier",
	1*256 + 40:  "8 Character Envelope Number",
	1*256 + 50:  "Product ID - Max 32 characters",
	1*256 + 60:  "Envelope Priority - 1 numeric characters",
	1*256 + 70:  "Date Sent - 8 numeric characters CCYYMMDD",
	1*256 + 80:  "Time Sent - 11 characters HHMMSS±HHMM",
	1*256 + 90:  "Coded Character Set - Max 32 characters",
	1*256 + 100: "UNO (Unique Name of Object) - 14 to 80 characters",
	1*256 + 120: "ARM Identifier - 2 byte binary number",
	1*256 + 122: "ARM Version - 2 byte binary number",

	// Application Record
	2*256 + 0:   "Record Version - 2 byte binary number",
	2*256 + 3:   "Object Type Reference -  3 plus 0 to 64 Characters",
	2*256 + 5:   "Object Name (Title) - Max 64 characters",
	2*256 + 7:   "Edit Status - Max 64 characters",
	2*256 + 8:   "Editorial Update - 2 numeric characters",
	2*256 + 10:  "Urgency - 1 numeric character",
	2*256 + 12:  "Subject Reference - 13 to 236 characters",
	2*256 + 15:  "Category - Max 3 characters",
	2*256 + 20:  "Supplemental Category - Max 32 characters",
	2*256 + 22:  "Fixture Identifier - Max 32 characters",
	2*256 + 25:  "Keywords - Max 64 characters",
	2*256 + 26:  "Content Location Code - 3 characters",
	2*256 + 27:  "Content Location Name - Max 64 characters",
	2*256 + 30:  "Release Date - 8 numeric characters CCYYMMDD",
	2*256 + 35:  "Release Time - 11 characters HHMMSS±HHMM",
	2*256 + 37:  "Expiration Date - 8 numeric characters CCYYMMDD",
	2*256 + 35:  "Expiration Time - 11 characters HHMMSS±HHMM",
	2*256 + 40:  "Special Instructions - Max 256 Characters",
	2*256 + 42:  "Action Advised - 2 numeric characters",
	2*256 + 45:  "Reference Service - Max 10 characters",
	2*256 + 47:  "Reference Date - 8 numeric characters CCYYMMDD",
	2*256 + 50:  "Reference Number - 8 characters",
	2*256 + 55:  "Date Created - 8 numeric characters CCYYMMDD",
	2*256 + 60:  "Time Created - 11 characters HHMMSS±HHMM",
	2*256 + 62:  "Digital Creation Date - 8 numeric characters CCYYMMDD",
	2*256 + 63:  "Digital Creation Time - 11 characters HHMMSS±HHMM",
	2*256 + 65:  "Originating Program - Max 32 characters",
	2*256 + 70:  "Program Version - Max 10 characters",
	2*256 + 75:  "Object Cycle - 1 character",
	2*256 + 80:  "By-Line (Author) - Max 32 Characters",
	2*256 + 85:  "By-Line Title (Author Position) - Max 32 characters",
	2*256 + 90:  "City - Max 32 Characters",
	2*256 + 92:  "Sub-Location - Max 32 characters",
	2*256 + 95:  "Province/State - Max 32 Characters",
	2*256 + 100: "Country/Primary Location Code - 3 alphabetic characters",
	2*256 + 101: "Country/Primary Location Name - Max 64 characters",
	2*256 + 103: "Original Transmission Reference - Max 32 characters",
	2*256 + 105: "Headline - Max 256 Characters",
	2*256 + 110: "Credit - Max 32 Characters",
	2*256 + 115: "Source - Max 32 Characters",
	2*256 + 116: "Copyright Notice - Max 128 Characters",
	2*256 + 118: "Contact - Max 128 characters",
	2*256 + 120: "Caption/Abstract - Max 2000 Characters",
	2*256 + 122: "Caption Writer/Editor - Max 32 Characters",
	2*256 + 125: "Rasterized Caption - 7360 bytes, 1 bit per pixel, 460x128pixel image",
	2*256 + 130: "Image Type - 2 characters",
	2*256 + 131: "Image Orientation - 1 alphabetic character",
	2*256 + 135: "Language Identifier - 2 or 3 aphabetic characters",
	2*256 + 150: "Audio Type - 2 characters",
	2*256 + 151: "Audio Sampling Rate - 6 numeric characters",
	2*256 + 152: "Audio Sampling Resolution - 2 numeric characters",
	2*256 + 153: "Audio Duration - 6 numeric characters",
	2*256 + 154: "Audio Outcue - Max 64 characters",
	2*256 + 200: "ObjectData Preview File Format - 2 byte binary number",
	2*256 + 201: "ObjectData Preview File Format Version - 2 byte binary number",
	2*256 + 202: "ObjectData Preview Data - Max 256000 binary bytes",

	// Pre-ObjectData Descriptor Record
	7*256 + 10: "Size Mode - 1 numeric character",
	7*256 + 20: "Max Subfile Size",
	7*256 + 90: "ObjectData Size Announced",
	7*256 + 95: "Maximum ObjectData Size",

	// ObjectData Record
	8*256 + 10: "Subfile",

	// Post ObjectData Descriptor Record
	9*256 + 10: "Confirmed ObjectData Size",
}

/******************************************************************************
* End of Global Variable:     IPTC_Entry_Descriptions
******************************************************************************/

/******************************************************************************
* Global Variable:      IPTC_File Formats
*
* Contents:     The names of the IPTC-NAA IIM File Formats for field 1:20
*
******************************************************************************/

var aIPTCFileFormats = []string{
	"No ObjectData",
	"IPTC-NAA Digital Newsphoto Parameter Record",
	"IPTC7901 Recommended Message Format",
	"Tagged Image File Format (Adobe/Aldus Image data)",
	"Illustrator (Adobe Graphics data)",
	"AppleSingle (Apple Computer Inc)",
	"NAA 89-3 (ANPA 1312)",
	"MacBinary II",
	"IPTC Unstructured Character Oriented File Format (UCOFF)",
	"United Press International ANPA 1312 variant",
	"United Press International Down-Load Message",
	"JPEG File Interchange (JFIF)",
	"Photo-CD Image-Pac (Eastman Kodak)",
	"Microsoft Bit Mapped Graphics File [*.BMP]",
	"Digital Audio File [*.WAV] (Microsoft & Creative Labs)",
	"Audio plus Moving Video [*.AVI] (Microsoft)",
	"PC DOS/Windows Executable Files [*.COM][*.EXE]",
	"Compressed Binary File [*.ZIP] (PKWare Inc)",
	"Audio Interchange File Format AIFF (Apple Computer Inc)",
	"RIFF Wave (Microsoft Corporation)",
	"Freehand (Macromedia/Aldus)",
	"Hypertext Markup Language - HTML (The Internet Society)",
	"MPEG 2 Audio Layer 2 (Musicom), ISO/IEC",
	"MPEG 2 Audio Layer 3, ISO/IEC",
	"Portable Document File (*.PDF) Adobe",
	"News Industry Text Format (NITF)",
	"Tape Archive (*.TAR)",
	"Tidningarnas Telegrambyrå NITF version (TTNITF DTD)",
	"Ritzaus Bureau NITF version (RBNITF DTD)",
	"Corel Draw [*.CDR]",
}

/******************************************************************************
* End of Global Variable:     IPTC_File Formats
******************************************************************************/

/******************************************************************************
* Global Variable:      ImageType_Names
*
* Contents:     The names of the colour components for IPTC-NAA IIM field 2:130
*
******************************************************************************/

var aImageTypeNames = map[string]string{
	"M": "Monochrome",
	"Y": "Yellow Component",
	"A": "Magenta Component",
	"C": "Cyan Component",
	"K": "Black Component",
	"R": "Red Component",
	"G": "Green Component",
	"B": "Blue Component",
	"T": "Text Only",
	"F": "Full colour composite, frame sequential",
	"L": "Full colour composite, line sequential",
	"P": "Full colour composite, pixel sequential",
	"S": "Full colour composite, special interleaving",
}
