default fileName = in;

fileName|
open-gzip|
as-lines|
decode-formeta|
encode-json(prettyprinting="true")|
write("stdout");
