// Package tagging implements helper functions for attaching ISIL to records.
//
// Example output, tag vs tagger.
//
// $ taskcat AIIntermediateSchema --date 2020-04-15 | head | ./span-tag -unfreeze $(taskoutput AMSLFilterConfigFreeze) 2> /dev/null | jq -rc '[.["finc.id"], .["x.labels"][]] | @tsv'
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTY1  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTc0  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTgy  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTkx  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTk5  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjAy  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjA1  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjA5  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjEz  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjE1  DE-105  DE-14   DE-15   DE-82   DE-Brt1 DE-Ch1  DE-D275 DE-Gla1 DE-Zi4  DE-Zwi2
//
// $ taskcat AIIntermediateSchema --date 2020-04-15 | head | ./span-tagger -debug -db amsl.db 2> /dev/null
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTY1  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTc0  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTgy  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTkx  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMTk5  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjAy  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjA1  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjA5  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjEz  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTI0MS9qb2hva2FucmkuNDkuMjE1  DE-105, DE-14, DE-15, DE-82, DE-Brt1, DE-Ch1, DE-D275, DE-Gla1, DE-Zi4, DE-Zwi2
//
package tagging
