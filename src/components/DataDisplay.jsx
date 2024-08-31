import React, { useState, useEffect } from 'react';
import ReviewForm from '../components/ReviewForm';

const DataDisplay = () => {
    const [data, setData] = useState([]);

    useEffect(() => {
        fetch('http://localhost:8080/api/reviews')
            .then(response => response.json())
            .then(data => {
                // Ensure data is an array before setting it
                if (Array.isArray(data)) {
                    setData(data);
                } else {
                    console.error('Expected data to be an array:', data);
                }
            })
            .catch(error => console.error('Error fetching data:', error));
    }, []);

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-3xl font-bold text-center mb-8">Reviews from my past clients</h1>
            {data && data.length > 0 ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {data.map(item => (
                        <div key={item.id} className="bg-white shadow-lg rounded-lg p-6">
                            <h2 className="text-xl font-semibold mb-2">{item.name}</h2>
                            <h3 className="text-lg text-gray-700 font-semibold mb-2">{item.company}</h3>
                            <p className="text-gray-700">{item.review}</p>
                        </div>
                    ))}
                </div>
            ) : (
                <p className="text-center text-gray-500">No Reviews found</p>
            )}
            <ReviewForm />
        </div>
    );
};

export default DataDisplay;
