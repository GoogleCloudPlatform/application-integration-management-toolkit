local f = import "functions";
local inputs = {"prefix": "test-","another-prefix": "hello-world"};
{"full-custom-header": (inputs["prefix"]+inputs["another-prefix"])}