import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '../css/addvendor.css';

function AddVendor() {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [subscriptionDays, setSubscriptionDays] = useState('');
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(null);
  const [imageError, setImageError] = useState(null);

  const [fieldErrors, setFieldErrors] = useState({
    name: false,
    description: false,
    image: false,
    subscriptionDays: false,
  });
  const [errorMessages, setErrorMessages] = useState({
    name: '',
    description: '',
    image: '',
    subscriptionDays: '',
    general: '',
  });
  const navigate = useNavigate();

  const handleSave = async (e) => {
    e.preventDefault();

    // Initialize error state
    let hasErrors = false;

    // Reset previous field errors and messages
    setFieldErrors({
      name: false,
      description: false,
      image: false,
      subscriptionDays: false,
    });
    setErrorMessages({
      name: '',
      description: '',
      image: '',
      subscriptionDays: '',
      general: '',
    });

    // Validate name
    if (!name.trim()) {
      setFieldErrors(prev => ({ ...prev, name: true }));
      setErrorMessages(prev => ({ ...prev, name: 'Name is required.' }));
      hasErrors = true;
    }

    // Validate description
    if (!description.trim()) {
      setFieldErrors(prev => ({ ...prev, description: true }));
      setErrorMessages(prev => ({ ...prev, description: 'Description is required.' }));
      hasErrors = true;
    }

    // Validate subscriptionDays only if it is not empty
    if (subscriptionDays.trim() && isNaN(subscriptionDays)) {
      setFieldErrors(prev => ({ ...prev, subscriptionDays: true }));
      setErrorMessages(prev => ({ ...prev, subscriptionDays: 'Valid subscription days are required.' }));
      hasErrors = true;
    }

    // Validate image
    if (imageError) {
      setFieldErrors(prev => ({ ...prev, image: true }));
      setErrorMessages(prev => ({ ...prev, image: imageError }));
      hasErrors = true;
    }

    if (hasErrors) {
      return; // Exit if there are validation errors
    }

    const formData = new FormData();
    formData.append('name', name);
    formData.append('description', description);
    formData.append('subscriptionDays', subscriptionDays);
    if (image) {
      formData.append('img', image);
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch('http://localhost:8080/vendors', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        if (errorData && errorData.error) {
          const errorObject = errorData.error;
          setFieldErrors({
            name: !!errorObject.name,
            description: !!errorObject.description,
            image: !!errorObject.img,
            subscriptionDays: !!errorObject.subscriptionDays,
          });
          setErrorMessages({
            name: errorObject.name || '',
            description: errorObject.description || '',
            image: errorObject.img || '',
            subscriptionDays: errorObject.subscriptionDays || '',
            general: '',
          });
        }
      } else {
        navigate('/vendors');
      }
    } catch (error) {
      console.error('Error adding vendor:', error);
      setErrorMessages(prev => ({ ...prev, general: 'An unknown error occurred.' }));
    }
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (!file) {
      setImage(null);
      setPreview(null);
      setImageError(null);
      return;
    }
    if (!file.type.match(/^image\/(png|jpg|jpeg|gif)$/)) {
      setImageError('Please select a valid image file');
      setImage(null);
      setPreview(null);
      return;
    }
    setImageError(null);
    setImage(file);

    const reader = new FileReader();
    reader.onloadend = () => {
      setPreview(reader.result);
    };
    if (file) {
      reader.readAsDataURL(file);
    }
  };

  const handleNameChange = (e) => {
    setName(e.target.value);
    if (e.target.value) {
      setFieldErrors(prev => ({ ...prev, name: false }));
      setErrorMessages(prev => ({ ...prev, name: '' }));
    }
  };

  const handleDescriptionChange = (e) => {
    setDescription(e.target.value);
    if (e.target.value) {
      setFieldErrors(prev => ({ ...prev, description: false }));
      setErrorMessages(prev => ({ ...prev, description: '' }));
    }
  };

  const handleSubscriptionDaysChange = (e) => {
    setSubscriptionDays(e.target.value);
    if (e.target.value && !isNaN(e.target.value)) {
      setFieldErrors(prev => ({ ...prev, subscriptionDays: false }));
      setErrorMessages(prev => ({ ...prev, subscriptionDays: '' }));
    }
  };

  return (
    <div className="vendor-container">
      <div className="edit-vendor-container">
        <div className="edit-vendor-image-container">
          <input
            type="file"
            id="image-upload"
            onChange={handleFileChange}
            style={{ display: 'none' }} // Hide the file input
          />
          <label htmlFor="image-upload" className="edit-vendor-image-label">
            {preview ? (
              <img src={preview} alt="Preview" className="vendor-image" />
            ) : (
              <div className="vendor-image"></div> // Removed 'No Image' text
            )}
          </label>
          {imageError && <p className="error-message">{imageError}</p>} {/* Move error message here */}
        </div>
        <div className="edit-vendor-info-container">
          <h1>Add Vendor</h1>
          {errorMessages.general && <p className="error-message">{errorMessages.general}</p>}
          <form onSubmit={handleSave}>
            <div className="form-group">
              <label htmlFor="name">Name</label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={handleNameChange}
                required
                style={{ borderColor: fieldErrors.name ? 'red' : '' }}
              />
              {errorMessages.name && <p className="error-message">{errorMessages.name}</p>}
            </div>
            <div className="form-group">
              <label htmlFor="description">Description</label>
              <textarea
                id="description"
                value={description}
                onChange={handleDescriptionChange}
                required
                style={{ borderColor: fieldErrors.description ? 'red' : '' }}
              />
              {errorMessages.description && <p className="error-message">{errorMessages.description}</p>}
            </div>
            <div className="form-group">
              <label htmlFor="subscriptionDays">Subscription Days (Optional)</label>
              <input
                type="number"
                id="subscriptionDays"
                value={subscriptionDays}
                onChange={handleSubscriptionDaysChange}
                style={{ borderColor: fieldErrors.subscriptionDays ? 'red' : '' }}
              />
              {errorMessages.subscriptionDays && <p className="error-message">{errorMessages.subscriptionDays}</p>}
            </div>
            <button type="submit">Save Vendor</button>
          </form>
        </div>
      </div>
    </div>
  );
}

export default AddVendor;
