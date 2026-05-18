import React, { useState, useEffect } from 'react';
import api from '../api';
import { X } from 'lucide-react';

const TaskModal = ({ isOpen, onClose, task, projects, onSave }) => {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('TODO');
  const [storyPoints, setStoryPoints] = useState(0);
  const [dueDate, setDueDate] = useState('');
  const [projectId, setProjectId] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (task) {
      setTitle(task.title || '');
      setDescription(task.description || '');
      setStatus(task.status || 'TODO');
      setStoryPoints(task.story_points || 0);
      setDueDate(task.due_date || '');
      setProjectId(task.project_id || 0);
    } else {
      setTitle('');
      setDescription('');
      setStatus('TODO');
      setStoryPoints(0);
      setDueDate(new Date().toISOString().split('T')[0]);
      setProjectId(0);
    }
    setError(null);
  }, [task, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!title.trim()) return;

    setIsSubmitting(true);
    setError(null);
    const payload = {
      title,
      description,
      status,
      story_points: parseInt(storyPoints) || 0,
      due_date: dueDate,
      project_id: parseInt(projectId) || 0,
      position: task ? task.position : 0 // Preserve position if editing
    };

    try {
      if (task && task.id) {
        // Edit Mode
        await api.put(`/tasks/${task.id}`, payload);
        onSave({ ...task, ...payload });
      } else {
        // Create Mode
        const res = await api.post('/tasks', payload);
        onSave(res.data);
      }
      setIsSubmitting(false);
      onClose();
    } catch (err) {
      console.error('Failed to save task', err);
      setError(err.response?.data?.message || err.message || 'Network error or timeout. Please try again.');
      setIsSubmitting(false); // Stop loading so user can retry, but show error
    }
  };

  const handleDelete = async () => {
    if (task && task.id) {
      if (window.confirm('Are you sure you want to delete this task? This cannot be undone.')) {
        try {
          await api.delete(`/tasks/${task.id}`);
          onSave(null); // Passing null can trigger a refresh
          onClose();
        } catch (err) {
          console.error('Failed to delete task', err);
        }
      }
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2>{task ? 'Edit Task' : 'Create Task'}</h2>
          <button onClick={onClose} className="icon-btn"><X size={20} /></button>
        </div>
        
        {error && (
          <div style={{ padding: '0.5rem 1rem', background: 'var(--danger)', color: 'white', borderRadius: '4px', margin: '0 1rem', fontSize: '0.9rem' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="modal-form">
          <div className="form-group">
            <label>Title</label>
            <input type="text" value={title} onChange={e => setTitle(e.target.value)} required autoFocus className="inline-input"/>
          </div>
          
          <div className="form-group">
            <label>Description</label>
            <textarea value={description} onChange={e => setDescription(e.target.value)} className="inline-input" rows="3"></textarea>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Status</label>
              <select value={status} onChange={e => setStatus(e.target.value)} className="inline-input">
                <option value="TODO">To Do</option>
                <option value="IN_PROGRESS">In Progress</option>
                <option value="COMPLETED">Completed</option>
                <option value="CANCELLED">Cancelled</option>
              </select>
            </div>
            
            <div className="form-group">
              <label>Project</label>
              <select value={projectId} onChange={e => setProjectId(e.target.value)} className="inline-input">
                <option value={0}>Unassigned</option>
                {projects.map(p => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Due Date</label>
              <input type="date" value={dueDate} onChange={e => setDueDate(e.target.value)} className="inline-input" />
            </div>
            
            <div className="form-group">
              <label>Story Points</label>
              <input type="number" min="0" value={storyPoints} onChange={e => setStoryPoints(e.target.value)} className="inline-input" />
            </div>
          </div>

          <div className="modal-actions" style={{ justifyContent: task ? 'space-between' : 'flex-end' }}>
            {task && (
              <button type="button" onClick={handleDelete} className="text-btn" style={{ color: 'var(--danger)' }}>
                Delete Task
              </button>
            )}
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button type="button" onClick={onClose} className="text-btn">Cancel</button>
              <button type="submit" className="primary-btn btn-small" disabled={isSubmitting}>
                {isSubmitting ? 'Saving...' : 'Save Task'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
