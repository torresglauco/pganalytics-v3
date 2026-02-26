#include "../include/thread_pool.h"

ThreadPool::ThreadPool(size_t num_threads)
    : stop_(false) {
    for (size_t i = 0; i < num_threads; ++i) {
        workers_.emplace_back(
            [this] {
                for (;;) {
                    std::function<void()> task;

                    {
                        std::unique_lock<std::mutex> lock(this->queue_mutex_);
                        this->condition_.wait(lock,
                            [this] { return this->stop_ || !this->tasks_.empty(); });

                        // Exit worker thread if pool is stopping and no more tasks
                        if (this->stop_ && this->tasks_.empty()) {
                            return;
                        }

                        task = std::move(this->tasks_.front());
                        this->tasks_.pop();
                    }

                    // Execute task outside the lock
                    task();
                }
            }
        );
    }
}

ThreadPool::~ThreadPool() {
    {
        std::unique_lock<std::mutex> lock(queue_mutex_);
        stop_ = true;
    }
    condition_.notify_all();

    // Wait for all threads to finish
    for (std::thread& worker : workers_) {
        worker.join();
    }
}
