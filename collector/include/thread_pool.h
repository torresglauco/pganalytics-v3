#pragma once

#include <thread>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <functional>
#include <memory>
#include <vector>
#include <future>
#include <stdexcept>

/**
 * Thread Pool for parallel execution of collector tasks
 *
 * Implements a fixed-size thread pool using worker threads that:
 * - Execute tasks from a queue in FIFO order
 * - Block when queue is empty
 * - Support graceful shutdown with pending task completion
 * - Provide future-based task result retrieval
 *
 * Usage:
 *   ThreadPool pool(4);  // 4 worker threads
 *   auto future = pool.enqueue([]() { return computeSomething(); });
 *   auto result = future.get();  // Wait for result
 */
class ThreadPool {
public:
    /**
     * Create a thread pool with specified number of worker threads
     * @param num_threads Number of worker threads to create
     */
    explicit ThreadPool(size_t num_threads);

    /**
     * Destructor - joins all worker threads after completing pending tasks
     */
    ~ThreadPool();

    /**
     * Enqueue a task for execution
     * @param f Task function/lambda to execute
     * @return Future for retrieving the result
     */
    template<class F>
    auto enqueue(F&& f) -> std::future<typename std::result_of<F()>::type> {
        using return_type = typename std::result_of<F()>::type;

        auto task = std::make_shared<std::packaged_task<return_type()>>(
            std::forward<F>(f)
        );

        std::future<return_type> res = task->get_future();
        {
            std::unique_lock<std::mutex> lock(queue_mutex_);

            // Don't allow enqueueing after stopping the pool
            if (stop_) {
                throw std::runtime_error("enqueue on stopped ThreadPool");
            }

            tasks_.emplace([task]() { (*task)(); });
        }
        condition_.notify_one();
        return res;
    }

    /**
     * Get current queue size
     */
    size_t getQueueSize() const {
        std::unique_lock<std::mutex> lock(queue_mutex_);
        return tasks_.size();
    }

    /**
     * Get number of worker threads
     */
    size_t getThreadCount() const {
        return workers_.size();
    }

private:
    // Need to keep track of threads so we can join them
    std::vector<std::thread> workers_;

    // The task queue
    std::queue<std::function<void()>> tasks_;

    // Synchronization
    mutable std::mutex queue_mutex_;
    std::condition_variable condition_;
    bool stop_;
};
