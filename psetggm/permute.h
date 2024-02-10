#ifdef __cplusplus
extern "C" {
#endif



void permute(unsigned int seed, uint32_t range, uint32_t* range_arr);
void invert_permutation(uint32_t* perm_array, uint32_t range, uint32_t* inv_array);
uint32_t fastMod(uint32_t x, uint32_t N);
uint32_t getRandIntModN(uint32_t N);


#ifdef __cplusplus
}
#endif
