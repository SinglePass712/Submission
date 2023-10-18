#include <iostream>
#include <ctime>
#include "../permute.h"

int main()
{
    int range = 32;
    std::cout << "range: " << range << std::endl;
    srand(time(NULL)); // randomize seed
    int start_s = clock();

    unsigned int* perm_arr = new unsigned int[range];
    permute(rand(),range, perm_arr);
    int stop_s = clock();
    std::cout << "time elapsed on permute: " <<((stop_s - start_s)/double(CLOCKS_PER_SEC)) << " seconds" << std::endl;

    for (int i = 0; i < range; i++)
    {
        std::cout << int(perm_arr[i]) << " ";
    }
    std::cout << std::endl;

    start_s = clock();
    unsigned int* inv_perm = new unsigned int[range];
    invert_permutation(perm_arr, range, inv_perm);
    stop_s = clock();
    std::cout << "time elapsed on invert: " <<((stop_s - start_s)/double(CLOCKS_PER_SEC)) << " seconds" << std::endl;

    // for (int i = 0; i < range; i++)
    // {
    //     std::cout << int(inv_perm[i]) << " ";
    // }
    // std::cout << std::endl;
}