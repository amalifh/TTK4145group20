// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;

// Note the return type: void*
void* incrementingThreadFunction(){
    for(int j = 0; i<1000000; j++){
        i++;
    }
   
    return NULL;
}

void* decrementingThreadFunction(){
    
    for(int j = 0; i<1000000; j++){
        i--;
    }
    return NULL;
}


int main(){
    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?

    //First we need to make the two threads so that we get a use of concurrency
    pthread_t incrementingThread;
    pthread_t decrementingThread;

    pthread_create(incrementingThread, NULL, incrementingThreadFunction(), NULL);
    pthread_create(decrementingThread, NULL, decrementingThreadFunction(), NULL);

    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    
    pthread_join(incrementingThread, i);
    pthread_join(decrementingThread, i);
 
    
    printf("The magic number is: %d\n", i);
    return 0;
}
