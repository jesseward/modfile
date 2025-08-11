# iMPULSE Amiga Mod Music Player

iMPULSE is a command line Amiga Mod (module) Music file player.

Module file (MOD music, tracker music) is a family of music file formats originating from the MOD file format on Amiga systems used in the late 1980s. Those who produce these files (using the software called music trackers) and listen to them form the worldwide MOD scene a part of the demoscene subculture. Protracker is a good example of such a player.

Module files store digitally recorded samples and several "patterns" or "pages" of music data in a form similar to that of a spreadsheet. These patterns contain note numbers, instrument numbers, and controller messages. The number of notes that can be played simultaneously depends on how many "tracks" there are per pattern. And the song is built of a pattern list, that tells in what order these patterns shall be played in the song.

The project is written in Go.

## Current Status

* Currently supports and plays the Amiga Protracker Music Format (Mod file) via the UI and command line


## References

* XM libary in C https://github.com/Artefact2/libxm
* Effect comparision between formats https://wiki.openmpt.org/Manual:_Effect_Reference
