#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Fast combined Answer function to save on allocations.
void answer(const uint8_t* pset, unsigned int pos, unsigned int univ_size, unsigned int set_size, unsigned int shift,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out);

//new fast answer for single pass pir

void answer_single_pass(uint8_t* db, unsigned int db_num_elems, unsigned int set_num_elems, unsigned int db_elem_size,
    uint8_t* parities, unsigned int seed_permutations, uint32_t* permutations, uint32_t* inverse_permutations);

#ifdef __cplusplus
} // extern "C" 
#endif

