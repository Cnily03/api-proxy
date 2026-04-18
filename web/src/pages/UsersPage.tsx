import {
  Button,
  Chip,
  Form,
  Input,
  Label,
  ListBox,
  ListBoxItem,
  Modal,
  Select,
  Table,
  TextField,
  toast,
} from "@heroui/react";
import { useCallback, useEffect, useState } from "react";
import type { User } from "@/utils/api";
import * as api from "@/utils/api";

type UserFormData = { username: string; password: string; role: string };

function UserForm({
  id,
  form,
  setForm,
  onSubmit,
  isEdit,
  forbidEnterSubmit = false,
}: {
  id: string;
  form: UserFormData;
  setForm: React.Dispatch<React.SetStateAction<UserFormData>>;
  onSubmit: () => void;
  isEdit?: boolean;
  forbidEnterSubmit?: boolean;
}) {
  return (
    <Form
      id={id}
      className="flex flex-col gap-4"
      validationBehavior="native"
      onKeyDown={
        forbidEnterSubmit
          ? (e) => {
              if (e.key === "Enter" && (e.target as HTMLElement).tagName !== "BUTTON") e.preventDefault();
            }
          : undefined
      }
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <TextField isRequired value={form.username} onChange={(v) => setForm((f) => ({ ...f, username: v }))}>
        <Label>用户名</Label>
        <Input placeholder="用户名" />
      </TextField>
      <TextField isRequired={!isEdit} value={form.password} onChange={(v) => setForm((f) => ({ ...f, password: v }))}>
        <Label>密码{isEdit ? "（留空则不修改）" : ""}</Label>
        <Input type="password" placeholder={isEdit ? "留空不修改" : "密码"} />
      </TextField>
      <Select
        isRequired
        selectedKey={form.role}
        onSelectionChange={(key) => setForm((f) => ({ ...f, role: String(key) }))}
      >
        <Label>角色</Label>
        <Select.Trigger>
          <Select.Value />
          <Select.Indicator />
        </Select.Trigger>
        <Select.Popover>
          <ListBox>
            <ListBoxItem id="user">普通用户</ListBoxItem>
            <ListBoxItem id="admin">管理员</ListBoxItem>
          </ListBox>
        </Select.Popover>
      </Select>
    </Form>
  );
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [form, setForm] = useState<UserFormData>({ username: "", password: "", role: "user" });

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      setUsers(await api.listUsers());
    } catch (err) {
      toast.danger(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function beginCreate() {
    setForm({ username: "", password: "", role: "user" });
    setCreateOpen(true);
  }

  function beginEdit(u: User) {
    setEditUser(u);
    setForm({ username: u.username, password: "", role: u.role });
  }

  async function doCreate() {
    await api.createUser(form.username, form.password, form.role);
    setCreateOpen(false);
    await loadData();
    toast.success("用户已创建");
  }

  async function doUpdate() {
    if (!editUser) return;
    const data: { username?: string; password?: string; role?: string } = {};
    if (form.username !== editUser.username) data.username = form.username;
    if (form.password) data.password = form.password;
    if (form.role !== editUser.role) data.role = form.role;
    await api.updateUser(editUser.id, data);
    setEditUser(null);
    await loadData();
    toast.success("用户已更新");
  }

  async function doDelete(id: number) {
    if (!window.confirm("确定删除该用户？")) return;
    await api.deleteUser(id);
    setUsers((prev) => prev.filter((u) => u.id !== id));
    toast.success("用户已删除");
  }

  return (
    <>
      <main className="max-w-5xl mx-auto px-4 py-6">
        <div className="flex items-center justify-between mb-3">
          <p className="uppercase tracking-wide">
            <span className="text-foreground text-xl font-semibold">用户管理</span>
            <span className="text-muted text-xs">（共 {users.length} 人）</span>
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onPress={() => void loadData()} isDisabled={loading}>
              <span className="icon-[mdi--refresh] w-4 h-4" />
              {loading ? "加载中..." : "刷新"}
            </Button>
            <Button variant="primary" size="sm" onPress={beginCreate}>
              <span className="icon-[mdi--account-plus-outline] w-4 h-4" />
              新增用户
            </Button>
          </div>
        </div>

        <Table aria-label="用户列表">
          <Table.Content>
            <Table.Header>
              <Table.Column isRowHeader>ID</Table.Column>
              <Table.Column>用户名</Table.Column>
              <Table.Column>角色</Table.Column>
              <Table.Column className="text-right">操作</Table.Column>
            </Table.Header>
            <Table.Body items={users} renderEmptyState={() => <span className="text-muted">暂无用户</span>}>
              {(u) => (
                <Table.Row id={u.id}>
                  <Table.Cell>#{u.id}</Table.Cell>
                  <Table.Cell className="font-medium">{u.username}</Table.Cell>
                  <Table.Cell>
                    <Chip size="sm">{u.role === "admin" ? "管理员" : "普通用户"}</Chip>
                  </Table.Cell>
                  <Table.Cell className="text-right">
                    <div className="flex gap-1.5 justify-end">
                      <Button variant="outline" size="sm" className="h-7 text-xs px-2" onPress={() => beginEdit(u)}>
                        <span className="icon-[mdi--pencil-outline] w-3 h-3" />
                        编辑
                      </Button>
                      <Button
                        variant="danger"
                        size="sm"
                        className="h-7 text-xs px-2"
                        onPress={() =>
                          void doDelete(u.id).catch((err) =>
                            toast.danger(err instanceof Error ? err.message : String(err))
                          )
                        }
                      >
                        <span className="icon-[mdi--delete-outline] w-3 h-3" />
                        删除
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              )}
            </Table.Body>
          </Table.Content>
        </Table>
      </main>

      {/* Create user modal */}
      <Modal>
        <Modal.Backdrop
          isOpen={createOpen}
          onOpenChange={(open) => {
            if (!open) setCreateOpen(false);
          }}
        >
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>新增用户</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <UserForm
                  id="form-create-user"
                  form={form}
                  setForm={setForm}
                  onSubmit={() => {
                    void doCreate().catch((err) => toast.danger(err instanceof Error ? err.message : String(err)));
                  }}
                  forbidEnterSubmit
                />
              </Modal.Body>
              <Modal.Footer>
                <Button variant="secondary" slot="close">
                  取消
                </Button>
                <Button type="submit" form="form-create-user" variant="primary">
                  创建
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>

      {/* Edit user modal */}
      <Modal>
        <Modal.Backdrop
          isOpen={editUser !== null}
          onOpenChange={(open) => {
            if (!open) setEditUser(null);
          }}
        >
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>编辑用户 {editUser?.username}</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <UserForm
                  id="form-edit-user"
                  form={form}
                  setForm={setForm}
                  onSubmit={() => {
                    void doUpdate().catch((err) => toast.danger(err instanceof Error ? err.message : String(err)));
                  }}
                  isEdit
                  forbidEnterSubmit
                />
              </Modal.Body>
              <Modal.Footer>
                <Button variant="secondary" slot="close">
                  取消
                </Button>
                <Button type="submit" form="form-edit-user" variant="primary">
                  保存
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  );
}
