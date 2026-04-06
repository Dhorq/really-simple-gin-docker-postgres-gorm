import { useState, useEffect } from 'react';
import { todoService, authService } from '../services/api';
import { useTodoStore } from '../store';
import { useAuthStore } from '../store';
import { Trash2, Pencil, Check, X } from 'lucide-react';
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

  const completedCount = todos.filter((t) => t.completed).length;

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-2xl mx-auto px-6 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-slate-800 tracking-tight">Taskly</h1>
            <p className="text-sm text-slate-500">{user?.email}</p>
          </div>
          <button
            onClick={handleLogout}
            className="text-sm text-slate-500 hover:text-slate-800 font-medium px-3 py-1.5 rounded-lg hover:bg-slate-100"
          >
            Sign out
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-2xl mx-auto px-6 py-8">
        {/* Add Form */}
        <form onSubmit={handleCreate} className="flex gap-2 mb-8">
          <input
            type="text"
            placeholder="What needs to be done?"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="flex-1 px-4 py-3 bg-white border border-slate-200 rounded-xl text-slate-800 placeholder-slate-400 focus:bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none"
            required
          />
          <button
            type="submit"
            disabled={loading}
            className="px-6 py-3 bg-blue-600 text-white font-semibold rounded-xl hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Add
          </button>
        </form>

        {/* Stats */}
        {todos.length > 0 && (
          <div className="flex items-center gap-2 mb-4">
            <span className="text-sm text-slate-500">
              {completedCount} of {todos.length} completed
            </span>
            <div className="flex-1 h-1 bg-slate-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-500 transition-all duration-300"
                style={{ width: `${todos.length ? (completedCount / todos.length) * 100 : 0}%` }}
              />
            </div>
          </div>
        )}

        {/* Todo List */}
        <ul className="space-y-2">
          {todos.map((todo) => (
            <li key={todo.id} className="bg-white rounded-xl border border-slate-200 overflow-hidden">
              {editingId === todo.id ? (
                <div className="p-4 space-y-4">
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    className="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-slate-800 focus:bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none"
                    required
                    autoFocus
                  />
                  <label className="flex items-center gap-3 cursor-pointer">
                    <div className={`relative w-10 h-6 rounded-full transition-colors ${editCompleted ? 'bg-blue-600' : 'bg-slate-300'}`}>
                      <input
                        type="checkbox"
                        checked={editCompleted}
                        onChange={(e) => setEditCompleted(e.target.checked)}
                        className="sr-only"
                      />
                      <div className={`absolute top-1 w-4 h-4 bg-white rounded-full shadow transition-transform ${editCompleted ? 'translate-x-5' : 'translate-x-1'}`} />
                    </div>
                    <span className="text-sm text-slate-600">Mark as complete</span>
                  </label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleSaveEdit(todo.id)}
                      disabled={loading}
                      className="flex items-center justify-center gap-2 flex-1 py-2 bg-blue-600 text-white font-semibold rounded-xl hover:bg-blue-700 disabled:opacity-50"
                    >
                      <Check size={16} />
                      Save
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="flex items-center justify-center gap-2 flex-1 py-2 text-slate-600 font-medium rounded-xl hover:bg-slate-100"
                    >
                      <X size={16} />
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-3 p-4">
                  <div
                    className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition-colors ${
                      todo.completed
                        ? 'bg-blue-600 border-blue-600'
                        : 'border-slate-300 hover:border-blue-400'
                    }`}
                  >
                    {todo.completed && (
                      <Check size={12} className="text-white" strokeWidth={3} />
                    )}
                  </div>
                  <span className={`flex-1 text-slate-800 ${todo.completed ? 'line-through text-slate-400' : ''}`}>
                    {todo.title}
                  </span>
                  <button
                    onClick={() => handleEditClick(todo)}
                    className="text-slate-400 hover:text-blue-600 p-2 rounded-lg hover:bg-blue-50 transition-colors"
                  >
                    <Pencil size={18} />
                  </button>
                  <button
                    onClick={() => handleDelete(todo.id)}
                    className="text-slate-400 hover:text-red-500 p-2 rounded-lg hover:bg-red-50 transition-colors"
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              )}
            </li>
          ))}
        </ul>

        {/* Empty State */}
        {todos.length === 0 && !loading && (
          <div className="text-center py-16">
            <div className="w-16 h-16 mx-auto mb-4 bg-slate-100 rounded-2xl flex items-center justify-center">
              <svg className="w-8 h-8 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3 className="text-slate-700 font-semibold mb-1">No tasks yet</h3>
            <p className="text-slate-500 text-sm">Add your first task above</p>
          </div>
        )}
      </main>
    </div>
  );
}
