SRC_DIR := ..
LIB := libpsetggm.a
HEADERS := $(addprefix $(SRC_DIR)/, pset_ggm.h)
SRC := $(addprefix $(SRC_DIR)/, pset_ggm.cpp Defines.cpp AES.cpp xor.cpp permute.cpp)
TESTS := test_pset test_distinct test_xor test_permute

OBJ := $(addsuffix .o,$(basename $(SRC)))

CCFLAGS := -std=c++11 -Ofast -I.. -g 

SIMD_FLAGS.x86_64 := -msse2 -msse -march=native -maes
SIMD_FLAGS.aarch64 := -march=armv8-a+fp+simd+crypto+crc


uname_p := $(strip $(shell uname -p))
FLAGS := $(CCFLAGS) $(SIMD_FLAGS.${uname_p}) -L/opt/homebrew/lib -lssl -lcrypto -I/opt/homebrew/include

all: $(LIB) $(TESTS)

test_pset: test_pset.cpp $(LIB)
	g++ $(FLAGS) -o $@ $^

test_distinct: test_distinct.cpp $(LIB)
	g++ $(FLAGS) -o $@ $^

test_xor: test_xor.cpp $(LIB)
	g++ $(FLAGS) -o $@ $^

test_permute: test_permute.cpp $(LIB)
	g++ $(FLAGS) -o $@ $^


%.o: %.cpp $(HEADERS)
	g++ $(FLAGS) -c $< -o $@ 

$(LIB): $(OBJ)
	ar rcs $@ $(OBJ)

clean: 
	rm -f main $(LIB) $(TESTS) $(OBJ)
