color-ssh
=========

Little bit of go to scratch an itch. Basically change the color of the terminal based upon the hostname we're sshing into.

We do this by creating a hash of the hostname and using that to select r, g, and b values for the background color. For
picking the foreground color we take the complement of this generated color. This isn't perfect as it can still result
in color schemes that are difficult to look at, but it's a start.
