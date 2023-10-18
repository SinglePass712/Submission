#include "answer.h"
#include <cmath>
#include "pset_ggm.h"
#include "xor.h"
#include "permute.h"
#include <iostream>
#include <vector>

extern "C" {

void answer(const uint8_t* pset, unsigned int pos, unsigned int univ_size, unsigned int set_size, unsigned int shift,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out) {
    auto worksize = workspace_size(univ_size, set_size+1);
    auto workspace = (uint8_t*)malloc(worksize+set_size*sizeof(long long unsigned int));
    auto gen = pset_ggm_init(univ_size, set_size+1, workspace);

    auto elems = (long long unsigned int*)(workspace+worksize);
    pset_ggm_eval_punc(gen, pset, pos, elems);

    for (int i = 0; i < set_size; i++) 
        elems[i] = ((elems[i]+shift)%univ_size)*row_len;


    xor_rows(db, db_len, elems, set_size, block_len, out);
    free(workspace);
}



//db_elem_size is in BYTES, size of parities array is db_elem_size*perm_size
// set_num_elems = set size is also equal to number of permutations
//here we assume that set_num_elems divides N

const __m128i one = _mm_setr_epi32(0, 0, 0, 1);

void answer_single_pass(uint8_t* db, unsigned int db_num_elems, unsigned int set_num_elems, unsigned int db_elem_size,
    uint8_t* parities, unsigned int seed_permutations, unsigned int* permutations, unsigned int* inverse_permutations) {


    srand(seed_permutations);
    unsigned int perm_size = db_num_elems/ set_num_elems;

    //note that size of permutation is also number of parities
    unsigned int* curr_permutation = permutations;
    unsigned int* curr_inverse = inverse_permutations;
    uint8_t* moving_db = db;
    for (int i = 0; i < set_num_elems; i++) { //iterate over each permutation (which pertains to a chunk of db)
        //use randomness from permutations, need to change this so that it is deterministic and recoverable
        
        
        permute(0, perm_size, curr_permutation);
        //short unsigned int* curr_inverse = new short unsigned int[perm_size];
        invert_permutation(curr_permutation, perm_size, curr_inverse); // possibly do this as part of permute function

        // uint32_t curr_place = i*perm_size;
        // for (int j = 0; j < perm_size; j++) {
        //     permutations[curr_place+j] = curr_permutation[j]; 
        // }
        xor_single_pass(moving_db, curr_inverse, perm_size, parities, db_elem_size);

        moving_db =  moving_db + (db_elem_size*perm_size); // adjust scope (partition of db that we are reading)
        curr_permutation = (curr_permutation+perm_size);
        curr_inverse = (curr_inverse+perm_size);;
    }

}

} // extern "C"