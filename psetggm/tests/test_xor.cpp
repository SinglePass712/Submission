#include <iostream>
#include "../xor.h"

int main()
{
    int row_len = 31;
    uint8_t db[row_len * 7];
    long long unsigned int idx1 = 3;
    long long unsigned int idx2 = 6;
    for (int i = 0; i < row_len; i++)
    {
        db[idx1*row_len + i] = 'X';
        db[idx2*row_len + i] = 'X' ^ (uint8_t)i;
    }

    long long unsigned int elems[] = {idx1*row_len, idx2*row_len};
    uint8_t out[row_len+3];
    xor_rows(db, sizeof(db), elems, 2, row_len, out+1);

    for (int i = 0; i < row_len; i++)
    {
        std::cout << int(out[1+i]) << " ";
    }
    std::cout << std::endl;




    std::cout << "----------new xor tests---------" << std::endl;
    uint8_t db2[32] = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1};
    uint8_t out2[16] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
    int elem_size = 16;
    int num_elems = 2;
    xor_all_rows(db2, 2,16,out2);
    std::cout << "out :" <<std::endl;
    for (int i = 0; i < sizeof(out2); i++) {
        std::cout << int(out2[i]) << " ";
    }
    std::cout << std::endl;
}