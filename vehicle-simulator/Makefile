# Vehicle Simulator - Simple Makefile with build/ directory
# Usage: make        - Build the project
#        make clean  - Remove build artifacts
#        make proto  - Regenerate protobuf files

# Build directory
BUILD_DIR = build

# Compiler settings
CXX = g++
CC = gcc
CXXFLAGS = -std=c++17 -Wall -Wextra -O2
CFLAGS = -Wall -O2

# Include paths
INCLUDES = -I/usr/share/nanopb -Igenerated/proto -Isrc

# Libraries (order matters - cpp lib first, then async C lib)
LDFLAGS = -lpaho-mqttpp3 -lpaho-mqtt3as -lpthread

# Source files
CPP_SRCS = src/main.cpp src/signals.cpp src/serialization.cpp src/transport.cpp
C_SRCS = generated/proto/vehicle.pb.c

# Nanopb sources (system path - compile directly)
NANOPB_SRCS = /usr/share/nanopb/pb_encode.c \
              /usr/share/nanopb/pb_decode.c \
              /usr/share/nanopb/pb_common.c

# Object files in build directory
CPP_OBJS = $(patsubst src/%.cpp,$(BUILD_DIR)/%.o,$(CPP_SRCS))
C_OBJS = $(BUILD_DIR)/vehicle.pb.o
OBJS = $(CPP_OBJS) $(C_OBJS)

# Target executable (in build directory)
TARGET = $(BUILD_DIR)/vehicle-simulator

# Default target - depends on protobuf generation
all: $(TARGET)
	@echo "Built: $(TARGET)"

# Create build directory
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# Link the executable
$(TARGET): $(OBJS) $(NANOPB_SRCS)
	@mkdir -p $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) -o $@ $(OBJS) $(NANOPB_SRCS) $(LDFLAGS)

# Compile C++ source files to build directory
$(BUILD_DIR)/%.o: src/%.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

# Compile generated protobuf C file
$(BUILD_DIR)/vehicle.pb.o: generated/proto/vehicle.pb.c | $(BUILD_DIR)
	$(CC) $(CFLAGS) $(INCLUDES) -Igenerated -c $< -o $@

# Generate protobuf files (only when proto file changes)
generated/proto/vehicle.pb.c generated/proto/vehicle.pb.h: proto/vehicle.proto
	@mkdir -p generated/proto
	python3 /usr/share/nanopb/generator/nanopb_generator.py --output-dir=generated proto/vehicle.proto
	@echo "Generated protobuf files"

# Manual protobuf regeneration target
proto:
	@mkdir -p generated/proto
	python3 /usr/share/nanopb/generator/nanopb_generator.py --output-dir=generated proto/vehicle.proto
	@echo "Generated protobuf files"

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)/
	@echo "Cleaned build artifacts"

# Clean everything including generated protobuf files
distclean: clean
	rm -rf generated/
	@echo "Cleaned all generated files"

.PHONY: all clean distclean proto
