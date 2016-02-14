package EXIF

import (
	"bufio"
	"encoding/binary"
	"io/ioutil"
	"os"
)

type segment struct {
	segType      byte
	segName      string
	segDesc      string
	segDataStart uint64
	segData      []byte
}

// jpegError is a trivial implementation of error
type jpegError struct {
	descr string
}

/******************************************************************************
*
* Function:     get_jpegHeader
*
* Description:  Reads all the JPEG header segments from an JPEG image file into an
*               array
*
* Parameters:   filename - the filename of the file to JPEG file to read
*
* Returns:      headerdata - Array of JPEG header segments
*               FALSE - if headers could not be read
*
******************************************************************************/

func getJPEGHeaderData(filename string) ([]segment, error) {

	segments := make([]segment)

	// Attempt to open the jpeg file - the at symbol supresses the error message about
	// not being able to open files. The file_exists would have been used, but it
	// does not work with files fetched over http or ftp.
	fi, err := os.Open(filename)

	// Check if the file opened successfully
	if err != nill {
		// Could't open the file - exit
		return segments, &jpegError{"Could not open file"}
	}

	reader := bufio.NewReader(fi)

	// Read the first two characters
	data = make([]byte, 2)
	reader.Read(data)

	// Check that the first two characters are 0xFF 0xDA  (SOI - Start of image)
	if data[0] != 0xFF || data[1] != 0xD8 {
		// No SOI (FF D8) at start of file - This probably isn't a JPEG file - close file and return;
		fi.Close()
		return segments, &jpegError{"This probably is not a JPEG file"}
	}

	// Read the third character
	reader.Read(data)

	// Check that the third character is 0xFF (Start of first segment header)
	if data[0] != 0xFF {
		// NO FF found - close file and return - JPEG is probably corrupted
		fi.Close()
		return segments, &jpegError{"No FF found, JPEG is probably corrupted"}
	}

	// Flag that we havent yet hit the compressed image data
	foundCompressedImageData := false

	// Cycle through the file until, one of: 1) an EOI (End of image) marker is hit,
	//                                       2) we have hit the compressed image data (no more headers are allowed after data)
	//                                       3) or end of file is hit

	// while ( ( $data{1} != "\xD9" ) && (! $foundCompressedImageData) && ( ! feof( $filehnd ) ))
	for data[1] != 0xD9 && !foundCompressedImageData {
		// Found a segment to look at.
		// Check that the segment marker is not a Restart marker - restart markers don't have size or data after them
		if data[1] < 0xD0 || data[1] > 0xD7 {
			// Segment isn't a Restart marker
			// Read the next two bytes (size)
			// sizestr = network_safe_fread( $filehnd, 2 );

			// convert the size bytes to an integer
			// $decodedsize = unpack ("nsize", $sizestr);
			var decodedSize uint16
			binary.Read(reader, binary.LittleEndian, &decodedSize)

			// Save the start position of the data
			segdatastart, err := file.Seek(os.SEEK_CUR, 0)

			// Read the segment data with length indicated by the previously read size
			segdata = make([]byte, decodedSize)
			reader.Read(segdata)

			// Store the segment information in the output array
			//$headerdata[] = array(  "SegType" : ord($data{1}),
			//                        "SegName" : $GLOBALS[ "JPEG_Segment_Names" ][ ord($data{1}) ],
			//                        "SegDesc" : $GLOBALS[ "JPEG_Segment_Descriptions" ][ ord($data{1}) ],
			//                        "SegDataStart" : $segdatastart,
			//                        "SegData" : $segdata );
			headerdata := &segment{}
			headerdata.segType = data[1]
			headerdata.segName = aJPEGSegmentNames[data[1]]
			headerdata.segDesc = aJPEGSegmentDescriptions[data[1]]
			headerdata.segDataStart = segdatastart
			headerdata.segData = segdata

			segments = append(segments, headerdata)
		}

		// If this is a SOS (Start Of Scan) segment, then there is no more header data - the compressed image data follows
		if data[1] == 0xDA {
			// Flag that we have hit the compressed image data - exit loop as no more headers available.
			foundCompressedImageData = true
		} else {
			// Not an SOS - Read the next two bytes - should be the segment marker for the next segment
			reader.Read(data)

			// Check that the first byte of the two is 0xFF as it should be for a marker
			if data[0] != 0xFF {
				// NO FF found - close file and return - JPEG is probably corrupted
				fi.Close()
				return segments, &jpegError{"No FF found, JPEG is probably corrupted"}
			}
		}
	}

	// Close File
	fi.Close()

	// Return the header data retrieved
	return segments, nil
}

/******************************************************************************
*
* Function:     put_jpegHeader
*
* Description:  Writes JPEG header data into a JPEG file. Takes an array in the
*               same format as from get_jpegHeader, and combines it with
*               the image data of an existing JPEG file, to create a new JPEG file
*               WARNING: As this function will replace all JPEG headers,
*                        including SOF etc, it is best to read the jpeg headers
*                        from a file, alter them, then put them back on the same
*                        file. If a SOF segment wer to be transfered from one
*                        file to another, the image could become unreadable unless
*                        the images were idenical size and configuration
*
*
* Parameters:   oldFilename - the JPEG file from which the image data will be retrieved
*               newFilename - the name of the new JPEG to create (can be same as oldFilename)
*               jpegHeader - a JPEG header data array in the same format
*                                  as from get_jpegHeader
*
* Returns:      TRUE - on Success
*               FALSE - on Failure
*
******************************************************************************/

func putJPEGHeaderData(oldFilename string, newFilename string, jpegHeader []segment) error {
	// Change: added check to ensure data exists, as of revision 1.10
	// Check if the data to be written exists

	// extract the compressed image data from the old file
	compressedImageData := get_jpeg_image_data(oldFilename)

	// Check if the extraction worked
	if compressedImageData == false || len(compressedImageData) == 0 {
		return &jpegError{"Couldn't get image data from old file"}
	}

	// Cycle through new headers
	// foreach ($jpegHeader as $segno : $segment)
	for _, segment := range jpegHeader {
		// Check that this header is smaller than the maximum size
		if len(segment.segData) > 0xfffd {
			return &jpegError{"A Header is too large to fit in JPEG segment"}
		}
	}

	// Attempt to open the new jpeg file
	fi, err := os.Open(newFilename)

	// Check if the file opened successfully
	if err != nill {
		return &jpegError{"Could not open file $newFilename"}
	}

	// Write SOI
	writer := bufio.NewWriter(fi)
	writer.Write(0xFF)
	writer.Write(0xD8)

	// Cycle through new headers, writing them to the new file
	// foreach ($jpegHeader as $segno : $segment)
	for _, seg := range jpegHeader {
		// Write segment marker
		// fwrite( $newfilehnd, sprintf( "\xFF%c", $segment['SegType'] ) );
		writer.Write(seg.segType)

		// Write segment size
		// fwrite( $newfilehnd, pack( "n", strlen($segment['SegData']) + 2 ) );
		writer.Write(uint16(len(seg.segData)))

		// Write segment data
		// fwrite( $newfilehnd, $segment['SegData'] );
		writer.Write(seg.segData)
	}

	// Write the compressed image data
	// fwrite( $newfilehnd, $compressedImageData );
	writer.Write(compressedImageData)

	// Write EOI
	// fwrite( $newfilehnd, "\xFF\xD9" );
	writer.Write(0xFF)
	writer.Write(0xD9)

	// Close File
	// fclose($newfilehnd);
	fi.Close()

	return nil
}

/******************************************************************************
*
* Function:     get_jpeg_Comment
*
* Description:  Retreives the contents of the JPEG Comment (COM = 0xFFFE) segment if one
*               exists
*
* Parameters:   jpegHeader - the JPEG header data, as retrieved
*                                  from the get_jpegHeader function
*
* Returns:      string - Contents of the Comment segement
*               FALSE - if the comment segment couldnt be found
*
******************************************************************************/

func getJPEGComment(jpegHeader []segment) (segment, error) {
	//Cycle through the header segments until COM is found or we run out of segments
	for _, seg := range jpegHeader {
		if seg.segName == "COM" {
			return seg, nil
		}
	}
	return nil, &jpegError{"Couldn't find comment segment"}
}

/******************************************************************************
*
* Function:     put_jpeg_Comment
*
* Description:  Creates a new JPEG Comment segment from a string, and inserts
*               this segment into the supplied JPEG header array
*
* Parameters:   jpegHeader - a JPEG header data array in the same format
*                                  as from get_jpegHeader, into which the
*                                  new Comment segment will be put
*               $new_Comment - a string containing the new Comment
*
* Returns:      jpegHeader - the JPEG header data array with the new
*                                  JPEG Comment segment added
*
******************************************************************************/

func putJPEGComment(jpegHeader []segment, newComment string) ([]segment, bool) {
	//Cycle through the header segments
	// for( $i = 0; $i < count( $jpegHeader ); $i++ )
	for i, seg := range jpegHeader {
		// If we find an COM header,
		if seg == "COM" {
			// Found a preexisting Comment block - Replace it with the new one and return.
			jpegHeader[i].segData = new_Comment.segData
			return jpegHeader, true
		}
	}

	// No preexisting Comment block found, find where to put it by searching for the highest app segment
	for i, seg := range jpegHeader {
		if seg.segType < 0xE0 {

			comment := &segment{}
			comment.segType = 0xFE
			comment.segName = aJPEGSegmentNames[0xFE]
			comment.segDesc = JPEG_Segment_Descriptions[0xFE]
			comment.segData = []byte(newComment)

			jpegHeader = append(s, 0)
			copy(jpegHeader[i+1:], jpegHeader[i:])
			jpegHeader[i] = comment
			return jpegHeader, true
		}
	}
	return jpegHeader, false
}

/******************************************************************************
*
* Function:     get_jpeg_image_data
*
* Description:  Retrieves the compressed image data part of the JPEG file
*
* Parameters:   filename - the filename of the JPEG file to read
*
* Returns:      compressedData - A string containing the compressed data
*               FALSE - if retrieval failed
*
******************************************************************************/

func getJPEGImageData(filename string) ([]byte, error) {

	// Attempt to open the jpeg file
	fi, err := os.Open(filename)

	// Check if the file opened successfully
	if err != nil {
		// Could't open the file - exit
		return nil, &jpegError{"Could not open the file"}
	}

	reader := bufio.NewReader(fi)

	// Read the first two characters
	data := make([]byte, 2)
	reader.Read(data)

	// Check that the first two characters are 0xFF 0xDA  (SOI - Start of image)
	if data[0] != 0xFF || data[1] != 0xD8 {
		// No SOI (FF D8) at start of file - close file and return;
		fi.Close()
		return nil, &jpegError{"No SOI (FF D8) at start of file"}
	}

	// Read the third character
	reader.Read(data)

	// Check that the third character is 0xFF (Start of first segment header)
	if data[0] != 0xFF {
		// NO FF found - close file and return
		fi.Close()
		return nil, &jpegError{"No FF found"}
	}

	// Flag that we havent yet hit the compressed image data
	foundCompressedImageData := false

	// Cycle through the file until, one of: 1) an EOI (End of image) marker is hit,
	//                                       2) we have hit the compressed image data (no more headers are allowed after data)
	//                                       3) or end of file is hit

	// while ( ( $data{1} != "\xD9" ) && (! $foundCompressedImageData) && ( ! feof( $filehnd ) ))
	for data[1] != 0xD9 && !foundCompressedImageData {
		// Found a segment to look at.
		// Check that the segment marker is not a Restart marker - restart markers don't have size or data after them
		if data[1] < 0xD0 || data[1] > 0xD7 {
			// convert the size bytes (2) to an integer
			var decodedSize uint16
			binary.Read(reader, binary.LittleEndian, &decodedSize)

			// Read the segment data with length indicated by the previously read size
			segdata := make([]byte, decodedsize-2)
			reader.Read(segdata)
		}

		// If this is a SOS (Start Of Scan) segment, then there is no more header data - the compressed image data follows
		if data[1] == 0xDA {
			// Flag that we have hit the compressed image data - exit loop after reading the data
			foundCompressedImageData = true

			// read the rest of the file in
			// Can't use the filesize function to work out
			// how much to read, as it won't work for files being read by http or ftp
			// So instead read 1Mb at a time till EOF

			compressedData, err := ioutil.ReadAll(reader)

			// Strip off EOI and anything after
			s := len(compressedData)
			// $EOI_pos = strpos( $compressedData, "\xFF\xD9" );
			if compressedData[s-2] == 0xFF && compressedData[s-1] == 0xD9 {
				compressedData = compressedData[0 : s-2]
			}
		} else {
			// Not an SOS - Read the next two bytes - should be the segment marker for the next segment
			reader.Read(data)

			// Check that the first byte of the two is 0xFF as it should be for a marker
			if data[0] != 0xFF {
				// Problem - NO FF found, close file and return";
				fi.Close()
				return nil, &jpegError{"No FF found"}
			}
		}
	}

	// Close File
	fi.Close()

	// Return the compressed data if it was found
	if foundCompressedImageData {
		return compressedData, nil
	}

	return nil, &jpegError{"No compressed data found"}
}

/******************************************************************************
* End of Function:     get_jpeg_image_data
******************************************************************************/

/******************************************************************************
* Global Variable:      JPEG_Segment_Names
*
* Contents:     The names of the JPEG segment markers, indexed by their marker number
*
******************************************************************************/

var aJPEGSegmentNames = map[byte]string{
	0xC0: "SOF0", 0xC1: "SOF1", 0xC2: "SOF2", 0xC3: "SOF4",
	0xC5: "SOF5", 0xC6: "SOF6", 0xC7: "SOF7", 0xC8: "JPG",
	0xC9: "SOF9", 0xCA: "SOF10", 0xCB: "SOF11", 0xCD: "SOF13",
	0xCE: "SOF14", 0xCF: "SOF15",
	0xC4: "DHT", 0xCC: "DAC",

	0xD0: "RST0", 0xD1: "RST1", 0xD2: "RST2", 0xD3: "RST3",
	0xD4: "RST4", 0xD5: "RST5", 0xD6: "RST6", 0xD7: "RST7",

	0xD8: "SOI", 0xD9: "EOI", 0xDA: "SOS", 0xDB: "DQT",
	0xDC: "DNL", 0xDD: "DRI", 0xDE: "DHP", 0xDF: "EXP",

	0xE0: "APP0", 0xE1: "APP1", 0xE2: "APP2", 0xE3: "APP3",
	0xE4: "APP4", 0xE5: "APP5", 0xE6: "APP6", 0xE7: "APP7",
	0xE8: "APP8", 0xE9: "APP9", 0xEA: "APP10", 0xEB: "APP11",
	0xEC: "APP12", 0xED: "APP13", 0xEE: "APP14", 0xEF: "APP15",

	0xF0: "JPG0", 0xF1: "JPG1", 0xF2: "JPG2", 0xF3: "JPG3",
	0xF4: "JPG4", 0xF5: "JPG5", 0xF6: "JPG6", 0xF7: "JPG7",
	0xF8: "JPG8", 0xF9: "JPG9", 0xFA: "JPG10", 0xFB: "JPG11",
	0xFC: "JPG12", 0xFD: "JPG13",

	0xFE: "COM", 0x01: "TEM", 0x02: "RES",
}

/******************************************************************************
* End of Global Variable:     JPEG_Segment_Names
******************************************************************************/

/******************************************************************************
* Global Variable:      JPEG_Segment_Descriptions
*
* Contents:     The descriptions of the JPEG segment markers, indexed by their marker number
*
******************************************************************************/

var aJPEGSegmentDescriptions = map[byte]string{
	/* JIF Marker byte pairs in JPEG Interchange Format sequence */
	0xC0: "Start Of Frame (SOF) Huffman  - Baseline DCT",
	0xC1: "Start Of Frame (SOF) Huffman  - Extended sequential DCT",
	0xC2: "Start Of Frame Huffman  - Progressive DCT (SOF2)",
	0xC3: "Start Of Frame Huffman  - Spatial (sequential) lossless (SOF3)",
	0xC5: "Start Of Frame Huffman  - Differential sequential DCT (SOF5)",
	0xC6: "Start Of Frame Huffman  - Differential progressive DCT (SOF6)",
	0xC7: "Start Of Frame Huffman  - Differential spatial (SOF7)",
	0xC8: "Start Of Frame Arithmetic - Reserved for JPEG extensions (JPG)",
	0xC9: "Start Of Frame Arithmetic - Extended sequential DCT (SOF9)",
	0xCA: "Start Of Frame Arithmetic - Progressive DCT (SOF10)",
	0xCB: "Start Of Frame Arithmetic - Spatial (sequential) lossless (SOF11)",
	0xCD: "Start Of Frame Arithmetic - Differential sequential DCT (SOF13)",
	0xCE: "Start Of Frame Arithmetic - Differential progressive DCT (SOF14)",
	0xCF: "Start Of Frame Arithmetic - Differential spatial (SOF15)",
	0xC4: "Define Huffman Table(s) (DHT)",
	0xCC: "Define Arithmetic coding conditioning(s) (DAC)",

	0xD0: "Restart with modulo 8 count 0 (RST0)",
	0xD1: "Restart with modulo 8 count 1 (RST1)",
	0xD2: "Restart with modulo 8 count 2 (RST2)",
	0xD3: "Restart with modulo 8 count 3 (RST3)",
	0xD4: "Restart with modulo 8 count 4 (RST4)",
	0xD5: "Restart with modulo 8 count 5 (RST5)",
	0xD6: "Restart with modulo 8 count 6 (RST6)",
	0xD7: "Restart with modulo 8 count 7 (RST7)",

	0xD8: "Start of Image (SOI)",
	0xD9: "End of Image (EOI)",
	0xDA: "Start of Scan (SOS)",
	0xDB: "Define quantization Table(s) (DQT)",
	0xDC: "Define Number of Lines (DNL)",
	0xDD: "Define Restart Interval (DRI)",
	0xDE: "Define Hierarchical progression (DHP)",
	0xDF: "Expand Reference Component(s) (EXP)",

	0xE0: "Application Field 0 (APP0) - usually JFIF or JFXX",
	0xE1: "Application Field 1 (APP1) - usually EXIF or XMP/RDF",
	0xE2: "Application Field 2 (APP2) - usually Flashpix",
	0xE3: "Application Field 3 (APP3)",
	0xE4: "Application Field 4 (APP4)",
	0xE5: "Application Field 5 (APP5)",
	0xE6: "Application Field 6 (APP6)",
	0xE7: "Application Field 7 (APP7)",

	0xE8: "Application Field 8 (APP8)",
	0xE9: "Application Field 9 (APP9)",
	0xEA: "Application Field 10 (APP10)",
	0xEB: "Application Field 11 (APP11)",
	0xEC: "Application Field 12 (APP12) - usually [picture info]",
	0xED: "Application Field 13 (APP13) - usually photoshop IRB / IPTC",
	0xEE: "Application Field 14 (APP14)",
	0xEF: "Application Field 15 (APP15)",

	0xF0: "Reserved for JPEG extensions (JPG0)",
	0xF1: "Reserved for JPEG extensions (JPG1)",
	0xF2: "Reserved for JPEG extensions (JPG2)",
	0xF3: "Reserved for JPEG extensions (JPG3)",
	0xF4: "Reserved for JPEG extensions (JPG4)",
	0xF5: "Reserved for JPEG extensions (JPG5)",
	0xF6: "Reserved for JPEG extensions (JPG6)",
	0xF7: "Reserved for JPEG extensions (JPG7)",
	0xF8: "Reserved for JPEG extensions (JPG8)",
	0xF9: "Reserved for JPEG extensions (JPG9)",
	0xFA: "Reserved for JPEG extensions (JPG10)",
	0xFB: "Reserved for JPEG extensions (JPG11)",
	0xFC: "Reserved for JPEG extensions (JPG12)",
	0xFD: "Reserved for JPEG extensions (JPG13)",

	0xFE: "Comment (COM)",
	0x01: "For temp private use arith code (TEM)",
	0x02: "Reserved (RES)",
}
