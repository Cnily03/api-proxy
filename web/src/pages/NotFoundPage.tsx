import { Button } from "@heroui/react";
import { useNavigate } from "react-router";
import NavBar from "@/components/NavBar";
import { useAuth } from "@/utils/auth";

export default function NotFoundPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  return (
    <div className="min-h-screen bg-background">
      {user ? <NavBar /> : null}
      <div className="min-h-[60vh] flex items-center justify-center p-4">
        <div className="flex flex-col items-center gap-4 text-center">
          <span className="icon-[mdi--compass-off-outline] w-16 h-16 text-muted" />
          <div className="flex flex-col gap-1">
            <h1 className="text-3xl font-semibold">404</h1>
            <p className="text-muted text-sm">页面不存在或已被移除</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="secondary" onPress={() => navigate(-1)}>
              返回上一页
            </Button>
            <Button variant="primary" onPress={() => navigate("/")}>
              回到首页
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
