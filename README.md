bittor
======

Parses bittorrent files. Atm it assumes there is a "master dictionary" with an info dictionary inside of it (This is what http://wiki.theory.org/BitTorrentSpecification#Metainfo_File_Structure sais). Dictionariy keys are assumed to only be strings, while values can be anything (http://wiki.theory.org/BitTorrentSpecification#Dictionaries). Will implement torrent encoding later when the main torrent client are done.