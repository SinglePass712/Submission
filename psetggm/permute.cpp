#include <cstdint>
#include <cstdio>
#include <cstring>
#include <algorithm>
#include "intrinsics.h"
#include <iostream>
#include "permute.h"
#include <openssl/rand.h>
//-L"/opt/homebrew/Cellar/libsodium/1.0.18_1/lib" -lsodium

extern "C"
{
    unsigned char buf[4];
    //use fisher yates algorithm to sample a permutation
   
    void permute(unsigned int seed, unsigned int range, unsigned int* range_arr)
    {
        
        
        for (unsigned int i = 0; i < range; i++) {
            range_arr[i] = i;
        }

        //permute array using fisher-yates
        for (unsigned int i = range-1; i > 0; i--){

            //change rand() to cryptographically safe randomness from AES (can use one AES call > once)
            unsigned int j = getRandIntModN(i+1);
            //std::cout << "random int: " << j << std::endl;
            std::swap(range_arr[i],range_arr[j]);

        }

    }


    void invert_permutation(unsigned int* perm_array, unsigned int range, unsigned int* inv_array) {
        //short unsigned int* inv_perm_arr = new short unsigned int[range];

        for (unsigned int i = 0; i < range; i++) {
            inv_array[perm_array[i]] = i;
        }



    }

    uint32_t getRandIntModN(uint32_t N) {
        RAND_bytes(buf,4);
        return fastMod(uint32_t((unsigned char)(buf[0]) << 24 |
            (unsigned char)(buf[1]) << 16 |
            (unsigned char)(buf[2]) << 8 |
            (unsigned char)(buf[3])),N);
    }

    //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
    uint32_t fastMod(uint32_t x, uint32_t N) {
        return ((uint64_t) x * (uint64_t) N) >> 32;
    }


} // extern "C"
