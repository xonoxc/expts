
# circular_queue — C11 + pthreads
CC       ?= gcc
AR       ?= ar

BUILD    := build
INCLUDE  := include
SRC      := src
TESTS    := tests
BENCH    := benchmarks

CPPFLAGS += -I$(INCLUDE)
CFLAGS   ?= -O2 -g
CFLAGS   += -std=c11 -Wall -Wextra -Wpedantic -pthread
LDFLAGS  += -pthread
LDLIBS   +=

ASAN_FLAGS := -fsanitize=address -fno-omit-frame-pointer
TSAN_FLAGS := -fsanitize=thread -fno-omit-frame-pointer

# Benchmark configuration
PRODUCERS ?= 1
CONSUMERS ?= 1

# Criterion
CRITERION_CFLAGS := $(shell pkg-config --cflags criterion)
CRITERION_LIBS   := $(shell pkg-config --libs criterion)

# Outputs
LIB         := $(BUILD)/lib/libqueue.a
QUEUE_OBJ   := $(BUILD)/obj/queue.o

TEST_OBJ    := $(BUILD)/obj/test_queue.o
TEST_BIN    := $(BUILD)/bin/test_queue

ASAN_DIR    := $(BUILD)/asan
ASAN_BIN    := $(ASAN_DIR)/bin/test_queue
ASAN_LIB    := $(ASAN_DIR)/lib/libqueue.a

TSAN_DIR    := $(BUILD)/tsan
TSAN_BIN    := $(TSAN_DIR)/bin/test_queue_tsan
TSAN_LIB    := $(TSAN_DIR)/lib/libqueue.a

BENCH_OBJ   := $(BUILD)/obj/bench_$(PRODUCERS)_$(CONSUMERS).o
BENCH_BIN   := $(BUILD)/bin/bench_$(PRODUCERS)_$(CONSUMERS)

.PHONY: all build test asan tsan bench bench-build clean run-benchmarks stress

all: build

build: $(LIB)

# Directories
$(BUILD)/obj $(BUILD)/lib $(BUILD)/bin:
	mkdir -p $@

# Static library
$(QUEUE_OBJ): $(SRC)/queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) -MMD -MP -c $< -o $@

$(LIB): $(QUEUE_OBJ)
	mkdir -p $(dir $@)
	$(AR) rcs $@ $^

# Unit tests
$(TEST_OBJ): $(TESTS)/test_queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) $(CRITERION_CFLAGS) -MMD -MP -c $< -o $@

$(TEST_BIN): $(TEST_OBJ) $(LIB)
	mkdir -p $(dir $@)
	$(CC) $(LDFLAGS) $^ $(CRITERION_LIBS) $(LDLIBS) -o $@

test: $(TEST_BIN)
	./$(TEST_BIN)

# AddressSanitizer
$(ASAN_DIR)/obj/queue.o: $(SRC)/queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) $(ASAN_FLAGS) -MMD -MP -c $< -o $@

$(ASAN_DIR)/obj/test_queue.o: $(TESTS)/test_queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) $(ASAN_FLAGS) $(CRITERION_CFLAGS) -MMD -MP -c $< -o $@

$(ASAN_LIB): $(ASAN_DIR)/obj/queue.o
	mkdir -p $(dir $@)
	$(AR) rcs $@ $^

$(ASAN_BIN): $(ASAN_DIR)/obj/test_queue.o $(ASAN_LIB)
	mkdir -p $(dir $@)
	$(CC) $(LDFLAGS) $(ASAN_FLAGS) $^ $(CRITERION_LIBS) $(LDLIBS) -o $@

asan: $(ASAN_BIN)
	./$(ASAN_BIN)


# ThreadSanitizer
$(TSAN_BIN): $(TESTS)/test_queue_sanitize.c $(SRC)/queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) $(TSAN_FLAGS) \
		$(TESTS)/test_queue_sanitize.c \
		$(SRC)/queue.c \
		$(LDFLAGS) $(TSAN_FLAGS) $(LDLIBS) -o $@

tsan: $(TSAN_BIN)
	./$(TSAN_BIN)

# Benchmarks
$(BENCH_OBJ): $(BENCH)/bench_queue.c $(INCLUDE)/queue.h
	mkdir -p $(dir $@)
	$(CC) $(CPPFLAGS) $(CFLAGS) \
		-DPRODUCERS=$(PRODUCERS) \
		-DCONSUMERS=$(CONSUMERS) \
		-MMD -MP -c $< -o $@

$(BENCH_BIN): $(BENCH_OBJ) $(LIB)
	mkdir -p $(dir $@)
	$(CC) $(LDFLAGS) $^ $(LDLIBS) -o $@

bench-build: $(BENCH_BIN)

bench: bench-build
	./$(BENCH_BIN)

# Include generated dependencies
-include $(QUEUE_OBJ:.o=.d)
-include $(TEST_OBJ:.o=.d)
-include $(BENCH_OBJ:.o=.d)
-include $(ASAN_DIR)/obj/*.d
-include $(TSAN_DIR)/obj/*.d

clean:
	rm -rf $(BUILD)


run-benchmarks:
	./scripts/run_benchmarks.sh

stress: $(TEST_BIN)
	./scripts/run_test_thousand_times.sh
