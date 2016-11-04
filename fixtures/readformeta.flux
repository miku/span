default fileName = in;

fileName|
open-file|
as-lines|
decode-formeta|
encode-json(prettyprinting="true")|
write("stdout");
