# Copyright 2013 Prometheus Team
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

MAKE_ARTIFACTS = dumper

all: dumper

SUFFIXES:

include ../../Makefile.INCLUDE

dumper: $(shell find . -iname '*.go')
	$(GO) build -o dumper .

clean:
	rm -rf $(MAKE_ARTIFACTS)

.PHONY: clean
