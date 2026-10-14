import api from ".";
export type LoginResponse = {
    access_token: string,
    message: string,
    refresh_token: string
}
export type UserResponse = {
    email: string
    id: number
    name: string
  }
  export type RegisterInput = {
    name: string
    email: string
    password: string
  }

  export const registerUser = async (body: RegisterInput) =>
    api
      .post<ServerResponse<LoginResponse>>("/auth/register", body)
      .then((res) => res.data.data)
export const loginUser = async (email: string, password: string) => api.post<ServerResponse<LoginResponse>>("/auth/signin", {
  email,
  password
}).then(res => res.data.data);
export const getUserDetails = async () => api.get<ServerResponse<UserResponse>>("/auth/user").then(res => res.data.data);
