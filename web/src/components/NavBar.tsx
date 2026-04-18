import { Button, Dropdown, FieldError, Form, Input, Label, Modal, TextField, toast } from "@heroui/react";
import { useState } from "react";
import { NavLink, useNavigate } from "react-router";
import * as api from "@/utils/api";
import { useAuth } from "@/utils/auth";

const linkClass = ({ isActive }: { isActive: boolean }) =>
  `inline-flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg transition-colors ${isActive ? "bg-border font-semibold text-foreground" : "text-muted hover:bg-surface-secondary"}`;

export default function NavBar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const isAdmin = user?.role === "admin";
  const [pwOpen, setPwOpen] = useState(false);
  const [oldPw, setOldPw] = useState("");
  const [newPw, setNewPw] = useState("");
  const [confirmPw, setConfirmPw] = useState("");
  const [pwLoading, setPwLoading] = useState(false);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const handleChangePassword = async () => {
    setPwLoading(true);
    try {
      await api.changePassword(oldPw, newPw);
      toast.success("密码已修改");
      setPwOpen(false);
      setOldPw("");
      setNewPw("");
      setConfirmPw("");
    } catch (err) {
      toast.danger(err instanceof Error ? err.message : String(err));
    } finally {
      setPwLoading(false);
    }
  };

  return (
    <>
      <header className="bg-surface border-b border-border">
        <div className="max-w-5xl mx-auto px-4 py-2 flex items-center justify-between">
          <nav className="flex items-center gap-1">
            <NavLink to="/" className={linkClass}>
              <span className="icon-[mdi--swap-horizontal] w-4 h-4" />
              转发规则
            </NavLink>
            {isAdmin ? (
              <NavLink to="/users" className={linkClass}>
                <span className="icon-[mdi--account-group-outline] w-4 h-4" />
                用户管理
              </NavLink>
            ) : null}
          </nav>
          <Dropdown>
            <Dropdown.Trigger>
              <button
                type="button"
                className="inline-flex items-center gap-1.5 text-sm text-muted hover:text-foreground cursor-pointer transition-colors px-2 py-1 rounded-lg hover:bg-surface-secondary"
              >
                <span className="icon-[mdi--account-circle-outline] w-4 h-4" />
                {user?.username}
                <span className="icon-[mdi--chevron-down] w-3.5 h-3.5" />
              </button>
            </Dropdown.Trigger>
            <Dropdown.Popover placement="bottom end">
              <Dropdown.Menu
                onAction={(key) => {
                  if (key === "password") {
                    setOldPw("");
                    setNewPw("");
                    setConfirmPw("");
                    setPwOpen(true);
                  }
                  if (key === "logout") handleLogout();
                }}
              >
                <Dropdown.Item id="password">
                  <span className="icon-[mdi--lock-outline] w-4 h-4 mr-1.5" />
                  修改密码
                </Dropdown.Item>
                <Dropdown.Item id="logout">
                  <span className="icon-[mdi--logout] w-4 h-4 mr-1.5" />
                  退出登录
                </Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown.Popover>
          </Dropdown>
        </div>
      </header>

      {/* Change password modal */}
      <Modal>
        <Modal.Backdrop
          isOpen={pwOpen}
          onOpenChange={(open) => {
            if (!open) setPwOpen(false);
          }}
        >
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>修改密码</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <Form
                  id="form-change-pw"
                  className="flex flex-col gap-4"
                  validationBehavior="native"
                  onSubmit={(e) => {
                    e.preventDefault();
                    void handleChangePassword();
                  }}
                >
                  <TextField isRequired value={oldPw} onChange={setOldPw}>
                    <Label>当前密码</Label>
                    <Input type="password" placeholder="当前密码" />
                  </TextField>
                  <TextField isRequired value={newPw} onChange={setNewPw}>
                    <Label>新密码</Label>
                    <Input type="password" placeholder="新密码" />
                  </TextField>
                  <TextField
                    isRequired
                    value={confirmPw}
                    onChange={setConfirmPw}
                    validate={(v) => (v !== newPw ? "两次输入的密码不一致" : null)}
                  >
                    <Label>确认新密码</Label>
                    <Input type="password" placeholder="再次输入新密码" />
                    <FieldError />
                  </TextField>
                </Form>
              </Modal.Body>
              <Modal.Footer>
                <Button variant="secondary" slot="close">
                  取消
                </Button>
                <Button type="submit" form="form-change-pw" variant="primary" isDisabled={pwLoading}>
                  {pwLoading ? "提交中..." : "修改密码"}
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  );
}
