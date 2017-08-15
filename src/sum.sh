#!/bin/bash

tr '{}' '  ' | \
  awk '
/^unique/ { ucnt += $2; usize += $3; utim += $4;
  xcnt += $7; xsize += $8; xtim += $9;
  dcnt += $12; dsize += $13; dtim += $14
}
END {
  cnt = ucnt + dcnt
  print "count: " xcnt " (" (xcnt * 100.0 / cnt) "), " (ucnt - xcnt) " (" ((ucnt - xcnt) * 100.0 / cnt) "), " dcnt " (" (dcnt * 100 / cnt) ") / " cnt;

  usize *= 0.000001;
  xsize *= 0.000001;
  dsize *= 0.000001;

  size = usize + dsize
  print "size:  " xsize " (" (xsize * 100.0 / size) "), " (usize - xsize) " (" ((usize - xsize) * 100.0 / size) "), " dsize " (" (dsize * 100 / size) ") / " size;

  utim *= 0.000000001;
  xtim *= 0.000000001;
  dtim *= 0.000000001;

  tim = utim + dtim
  print "time:  " xtim " (" (xtim * 100.0 / tim) "), " (utim - xtim) " (" ((utim - xtim) * 100.0 / tim) "), " dtim " (" (dtim * 100 / tim) ") / " tim;
}
'
