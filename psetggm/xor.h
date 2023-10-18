#ifdef __cplusplus
extern "C" {
#endif


void xor_into(uint8_t* out, uint8_t* in, unsigned int elem_size);

void xor_rows(const uint8_t* db, unsigned int db_len, 
    const long long unsigned int* elems, unsigned int num_elems, 
    unsigned int block_len, uint8_t* out);

void xor_single_pass(uint8_t* db, uint32_t* inverse_perm, unsigned int perm_size,  uint8_t* parities, unsigned int elem_size);

void xor_all_rows(uint8_t* db, unsigned int num_elems,
        unsigned int elem_size, uint8_t* out);

void xor_hashes_by_bit_vector(const uint8_t* db, unsigned int db_len, 
    const uint8_t* indexing, 
    uint8_t* out);

#ifdef __cplusplus
}
#endif
