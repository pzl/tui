/*
ansi is a fairly simple wrapper around terminal escape codes.
So you don't have to remember that \033[2J clears the screen.
*/
package ansi

/*
--- Movement
<n>A-D move up,down,left,right
H - cursor to upper left (origin)
<y>;<x>H - move cursor to y,x
G - cursor to col 0
<n>G - cursor to col n

-- Clearing
K - clear line from cursor onwards
1K - clear from cursor left
2K - clear line cursor is on
J - clear screen from cursor down
1J - clear from cursor up
2J - clear the whole screen

--- Cursor
?25l - hide cursor
?25h show cursor
6n - get cursor coords
s - save position
u - restore position

--- screen
?1049h smcup
?1049l rmcup


--- color
-- 3/4-bit color (8,16 colors)
3<n>m -- blk,r,g,y,bl,mag,cy,w
3<n>;1m - bold version (or 9<n> for bright)
4<n>m - background

-- 8-bit color (256 colors)
38;5;<n>m - first 16 as above, then look up table
48;5;<n>m - for background

-- 24-bit color (16M colors)    $COLORTERM=truecolor || 24bit
38;2;<r>;<g>;<b>m - fg, 0-255
48;2;<r>;<g>;<b>m - bg, 0-255


--- text effects
0m - reset
1m - bold
2m - dim
3m - italic
4m - underline
5m - blink
7m - reverse bg/fg
8m - hidden (like for passwords, not *)
9m - strikethrough


?1000h - enable mouse
?1000l - disable mouse
*/
