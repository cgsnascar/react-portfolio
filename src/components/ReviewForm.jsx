import React, { useState } from 'react';

const ReviewForm = () => {
    const [company, setCompany] = useState('');
    const [name, setName] = useState('');
    const [review, setReview] = useState('');
    const [key, setKey] = useState('');
    const [message, setMessage] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();

        try {
            const response = await fetch('http://localhost:8080/api/review', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ company, name, review, key }),
            });

            if (response.ok) {
                setMessage('Review submitted successfully!');
                setCompany('');
                setName('');
                setReview('');
                setKey('');
            } else {
                setMessage('Failed to submit review. Please try again later.');
            }
        } catch (error) {
            console.error('Error submitting review:', error);
            setMessage('An error occurred. Please try again later.');
        }
    };

    return (
        <div className="max-w-md mx-auto mt-10 p-4 border border-gray-300 rounded-lg shadow-md">
            <h2 className="text-xl font-semibold mb-4">Submit a Review</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
                <div className="flex flex-col">
                    <label htmlFor="company" className="text-sm font-medium text-gray-700">Company</label>
                    <input
                        type="text"
                        id="company"
                        value={company}
                        onChange={(e) => setCompany(e.target.value)}
                        required
                        className="mt-1 p-2 border border-gray-300 rounded-md"
                    />
                </div>

                <div className="flex flex-col">
                    <label htmlFor="name" className="text-sm font-medium text-gray-700">Name</label>
                    <input
                        type="text"
                        id="name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                        className="mt-1 p-2 border border-gray-300 rounded-md"
                    />
                </div>

                <div className="flex flex-col">
                    <label htmlFor="review" className="text-sm font-medium text-gray-700">Review</label>
                    <textarea
                        id="review"
                        value={review}
                        onChange={(e) => setReview(e.target.value)}
                        required
                        rows="4"
                        className="mt-1 p-2 border border-gray-300 rounded-md"
                    />
                </div>

                <div className="flex flex-col">
                    <label htmlFor="key" className="text-sm font-medium text-gray-700">API Key</label>
                    <input
                        type="text"
                        id="key"
                        value={key}
                        onChange={(e) => setKey(e.target.value)}
                        required
                        className="mt-1 p-2 border border-gray-300 rounded-md"
                    />
                </div>

                <button
                    type="submit"
                    className="w-full py-2 px-4 bg-blue-500 text-white font-semibold rounded-md hover:bg-blue-600"
                >
                    Submit Review
                </button>

                {message && (
                    <p className={`mt-4 ${message.includes('successfully') ? 'text-green-600' : 'text-red-600'}`}>
                        {message}
                    </p>
                )}
            </form>
        </div>
    );
};

export default ReviewForm;
