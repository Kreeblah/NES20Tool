NES20Tool
=========

This tool is intended to read NES 2.0 headers and generate an XML file reflecting the syntactic meaning of the headers, as well as to take an XML file in the same format and apply it to a ROM set.

The tool uses the SHA256 hash of a ROM to determine which ROM the file contains, ignoring any existing iNES or NES 2.0 header currently on the ROM (if any; it also works with headerless ROMs for applying headers).  Other hashes are calculated and provided in generated XML files for convenience, but they have no significant meaning within this application.

Warning
-------

Even though the NES 2.0 header format has been around for a while, there still is not a definitive set of data to properly classify every ROM.  As such, applying a set of NES 2.0 headers to a ROM set is highly likely to cause some subset of the ROMs in the set to fail to work properly.

I don't presently intend to make formal releases until this situation settles out.

Known Issues and Potential Issues
---------------------------------

Some NES 2.0 header sets assign the value of "2" to byte 13 of the header for _some_ ROMs identified in byte 7 as PlayChoice 10 ROMs (the lower two bits as "2" in byte 7), but zero for this byte in others.  The reason for this is unknown, but the specification at https://wiki.nesdev.com/w/index.php/NES_2.0 implies that these ROMs should be assigned a value of "0" in byte 13 when those bits in byte 7 are not 1 or 3, so that's what this tool does in those circumstances.  Because of this, this tool is unable to model the entirety of those sets, and will result in differing data for headers for those ROMs (the value of byte 13 on those ROMs will be 0) if read to an XML file and re-applied to the same set.

Usage
-----

To use this tool, compile it for your favorite OS and then run it with the following options:

    -operation  
    	Operation to perform on the ROM set.  {read|write}  
    -rom-path string  
    	The path to a directory with NES ROMs to use for the operation.  
    -xml-file string  
    	The path to an XML file to use for the operation.
