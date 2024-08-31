import React, { useState } from 'react';

const ReviewForm = () => {
    const [name, setName] = useState('');
    const [company, setCompany] = useState('');
    const [review, setReview] = useState('');
    const [key, setKey] = useState('');
    const [message, setMessage] = useState('');

    const handleSubmit = async (event) => {
        event.preventDefault();

        // Perform form validation if needed
        if (!name || !company || !review || !key) {
            setMessage('All fields are required.');
            return;
        }

        try {
            const response = await fetch('http://localhost:8080/api/submit-review', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                    key: key,
                    name: name,
                    company: company,
                    review: review,
                }),
            });

            if (!response.ok) {
                throw new Error('Network response was not ok.');
            }

            const result = await response.text();
            setMessage('Review submitted successfully!');
            // Clear form fields after submission
            setName('');
            setCompany('');
            setReview('');
            setKey('');
        } catch (error) {
            setMessage('Failed to submit review. Please try again later.');
            console.error('Error submitting review:', error);
        }
    };

    return (
        <div className="container mx-auto p-4">
            <h2 className="text-3xl font-bold text-center mb-8">Submit a New Review</h2>
            <form onSubmit={handleSubmit} className="bg-white shadow-md rounded-lg p-4 max-w-md mx-auto">
                <div className="mb-3">
                    <label className="block text-gray-700 font-medium mb-1" htmlFor="name">Name:</label>
                    <input
                        type="text"
                        id="name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full border border-gray-300 p-2 rounded-sm text-sm"
                    />
                </div>
                <div className="mb-3">
                    <label className="block text-gray-700 font-medium mb-1" htmlFor="company">Company:</label>
                    <input
                        type="text"
                        id="company"
                        value={company}
                        onChange={(e) => setCompany(e.target.value)}
                        className="w-full border border-gray-300 p-2 rounded-sm text-sm"
                    />
                </div>
                <div className="mb-3">
                    <label className="block text-gray-700 font-medium mb-1" htmlFor="review">Review:</label>
                    <textarea
                        id="review"
                        value={review}
                        onChange={(e) => setReview(e.target.value)}
                        className="w-full border border-gray-300 p-2 rounded-sm text-sm"
                        rows="4"
                    ></textarea>
                </div>
                <div className="mb-3">
                    <label className="block text-gray-700 font-medium mb-1" htmlFor="key">Secret Key:</label>
                    <input
                        type="text"
                        id="key"
                        value={key}
                        onChange={(e) => setKey(e.target.value)}
                        className="w-full border border-gray-300 p-2 rounded-sm text-sm"
                    />
                </div>
                <button type="submit" className="bg-blue-500 text-white py-2 px-4 rounded-sm text-sm">
                    Submit
                </button>
                {message && <p className="mt-4 text-center text-red-500 text-sm">{message}</p>}
            </form>
        </div>
    );
};

export default ReviewForm;
