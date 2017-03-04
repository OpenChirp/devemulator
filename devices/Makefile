# Find all directories with .proto files and adds it's .pb.go target
DEVICE_PB_GO=$(subst .proto,.pb.go, $(wildcard */*.proto))

.PHONY: all clean

all: $(DEVICE_PB_GO)

%.pb.go: %.proto
	protoc --go_out=. $^

clean:
	$(RM) */*.pb.go
