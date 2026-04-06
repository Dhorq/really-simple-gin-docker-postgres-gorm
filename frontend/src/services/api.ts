import axios from 'axios';
import type { ApiResponse, LoginInput, RegisterInput, Todo, CreateTodoInput, UpdateTodoInput } from '../types';

const api = axios.create({
  baseURL: 'http://localhost:3000/api',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const authService = {
  register: (input: RegisterInput) =>
    api.post<ApiResponse<{ user: { id: number; email: string } }>>('/auth/register', input),

  login: (input: LoginInput) =>
    api.post<ApiResponse<{ user: { id: number; email: string } }>>('/auth/login', input),

  logout: () =>
    api.post<ApiResponse<void>>('/auth/logout'),

  me: () =>
    api.get<ApiResponse<{ id: number; email: string }>>('/auth/me'),
};

export const todoService = {
  getAll: () =>
    api.get<ApiResponse<Todo[]>>('/todos'),

  get: (id: number) =>
    api.get<ApiResponse<Todo>>(`/todos/${id}`),

  create: (input: CreateTodoInput) =>
    api.post<ApiResponse<Todo>>('/todos', input),

  update: (id: number, input: UpdateTodoInput) =>
    api.put<ApiResponse<Todo>>(`/todos/${id}`, input),

  delete: (id: number) =>
    api.delete<ApiResponse<void>>(`/todos/${id}`),
};
