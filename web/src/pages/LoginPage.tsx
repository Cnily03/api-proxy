"use client";

import { Button, Card, Input, Label, toast } from "@heroui/react";
import { useRef, useState } from "react";
import { useNavigate } from "react-router";
import * as api from "@/utils/api";
import { useAuth } from "@/utils/auth";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const enterHeld = useRef(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async () => {
    setLoading(true);
    try {
      await api.login(username, password);
      login();
      navigate("/");
    } catch (err) {
      if (err instanceof api.APIError) {
        if (err.response?.status === 429) {
          toast.danger("尝试次数过多，请稍后再试");
        } else if (err.response?.status === 401) {
          toast.danger("用户名或密码错误");
        } else {
          toast.danger(err.message);
        }
      } else {
        toast.danger(err instanceof Error ? err.message : String(err));
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <Card.Content className="p-6">
          <h1 className="text-xl font-semibold mb-6 text-center">用户登录</h1>
          <form
            className="flex flex-col gap-4"
            onSubmit={(e) => {
              e.preventDefault();
              void handleSubmit();
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                if (enterHeld.current) {
                  e.preventDefault();
                  return;
                }
                enterHeld.current = true;
              }
            }}
            onKeyUp={(e) => {
              if (e.key === "Enter") {
                enterHeld.current = false;
              }
            }}
          >
            <div className="flex flex-col gap-1">
              <Label htmlFor="login-user">用户名</Label>
              <Input
                id="login-user"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label htmlFor="login-pass">密码</Label>
              <Input
                id="login-pass"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
              />
            </div>
            <Button
              type="submit"
              variant="primary"
              fullWidth
              className="mt-2"
              isDisabled={loading || !username || !password}
            >
              {loading ? "登录中" : "登 录"}
            </Button>
          </form>
        </Card.Content>
      </Card>
    </div>
  );
}
