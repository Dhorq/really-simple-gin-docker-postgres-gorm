import { useState, useEffect } from 'react';
import { todoService, authService } from '../services/api';
import { useTodoStore } from '../store';
import { useAuthStore } from '../store';
import type { Todo } from '../types';

export default function TodoPage() {
  const { todos, setTodos, addTodo, updateTodo, removeTodo, setLoading, setError } = useTodoStore();
  const { user, logout } = useAuthStore();
  const [title, setTitle] = useState('');
  const [loading, setLocalLoading] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [editCompleted, setEditCompleted] = useState(false);

  useEffect(() => {
  const fetchTodos = async () => {
    setLoading(true);
    try {
      const res = await todoService.getAll();
      setTodos(res.data.data);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  fetchTodos();
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, []);


  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setLocalLoading(true);
    try {
      const res = await todoService.create({ title });
      addTodo(res.data.data);
      setTitle('');
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleEditClick = (todo: Todo) => {
    setEditingId(todo.id);
    setEditTitle(todo.title);
    setEditCompleted(todo.completed);
  };

  const handleSaveEdit = async (id: number) => {
    setLocalLoading(true);
    try {
      const res = await todoService.update(id, {
        title: editTitle,
        completed: editCompleted,
      });
      updateTodo(id, res.data.data);
      setEditingId(null);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditTitle('');
    setEditCompleted(false);
  };

  const handleDelete = async (id: number) => {
    try {
      await todoService.delete(id);
      removeTodo(id);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleLogout = async () => {
    try {
      await authService.logout();
      logout();
    } catch {
      logout();
    }
  };

  return (
    <div className="min-h-screen bg-slate-100 p-6">
      <div className="max-w-2xl mx-auto">
        <header className="flex justify-between items-center bg-white p-4 rounded-xl shadow-sm mb-6">
          <h2 className="text-blue-900 font-semibold">Hello, {user?.email}</h2>
          <button
            onClick={handleLogout}
            className="border-2 border-blue-600 text-blue-600 font-semibold px-4 py-2 rounded-lg hover:bg-blue-50 transition"
          >
            Logout
          </button>
        </header>

        <form onSubmit={handleCreate} className="flex gap-3 mb-6">
          <input
            type="text"
            placeholder="Add new todo..."
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="flex-1 border-2 border-blue-100 rounded-lg p-3 text-base focus:outline-none focus:border-blue-500"
            required
          />
          <button
            type="submit"
            disabled={loading}
            className="bg-blue-600 text-white font-semibold px-6 py-3 rounded-lg hover:bg-blue-700 transition disabled:bg-blue-300"
          >
            {loading ? 'Adding...' : 'Add'}
          </button>
        </form>

        <ul className="space-y-3">
          {todos.map((todo) => (
            <li key={todo.id} className="bg-white p-4 rounded-xl shadow-sm">
              {editingId === todo.id ? (
                <div className="space-y-3">
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    className="w-full border-2 border-blue-100 rounded-lg p-3 text-base focus:outline-none focus:border-blue-500"
                    required
                  />
                  <div className="flex items-center gap-4">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={editCompleted}
                        onChange={(e) => setEditCompleted(e.target.checked)}
                        className="w-5 h-5 accent-blue-600 cursor-pointer"
                      />
                      <span className="text-slate-600">Completed</span>
                    </label>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleSaveEdit(todo.id)}
                      disabled={loading}
                      className="flex-1 bg-blue-600 text-white font-semibold py-2 rounded-lg hover:bg-blue-700 transition disabled:bg-blue-300"
                    >
                      Save
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="flex-1 border-2 border-slate-300 text-slate-600 font-semibold py-2 rounded-lg hover:bg-slate-50 transition"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3 flex-1">
                    <button
                      onClick={() => handleEditClick(todo)}
                      className="text-blue-600 hover:text-blue-800 font-medium"
                    >
                      {todo.title}
                    </button>
                    {todo.completed && (
                      <span className="bg-green-100 text-green-700 text-xs font-semibold px-2 py-1 rounded-full">
                        Done
                      </span>
                    )}
                  </div>
                  <button
                    onClick={() => handleDelete(todo.id)}
                    className="bg-red-100 text-red-600 font-medium px-3 py-1 rounded-lg hover:bg-red-200 transition text-sm ml-4"
                  >
                    Delete
                  </button>
                </div>
              )}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
