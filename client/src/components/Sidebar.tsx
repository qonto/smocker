import * as React from "react";
import { connect } from "react-redux";
import { AppState } from "~modules/reducers";
import { Dispatch } from "redux";
import { Actions, actions } from "~modules/actions";
import { Button, Form, Input, Layout, Menu, Popover, Row, Spin } from "antd";
import "./Sidebar.scss";
import { Sessions, Session } from "~modules/types";

const NewButton = ({ onValidate }: any) => {
  const [visible, setVisible] = React.useState(false);
  const [name, setName] = React.useState("");
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onValidate(name.trim());
    setName("");
    setVisible(false);
  };
  const onChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setName(event.target.value);
  };
  return (
    <Popover
      placement="right"
      visible={visible}
      onVisibleChange={setVisible}
      content={
        <Form layout="inline" onSubmit={onSubmit}>
          <Form.Item>
            <Input value={name} onChange={onChange} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              Start
            </Button>
          </Form.Item>
        </Form>
      }
      title="You can set a name for the new session"
      trigger="click"
    >
      <Row align="middle" justify="center" type="flex">
        <Button ghost type="primary" icon="plus" className="session-button">
          New Session
        </Button>
      </Row>
    </Popover>
  );
};

interface Props {
  sessions: Sessions;
  loading: boolean;
  selected: string;
  fetch: () => void;
  selectSession: (sessionID: string) => void;
  newSession: (name: string) => void;
}

const SideBar = ({
  fetch,
  selected,
  sessions,
  loading,
  selectSession,
  newSession
}: Props) => {
  React.useLayoutEffect(() => {
    fetch();
  }, []);
  if (!selected && sessions.length > 0) {
    selectSession(sessions[0].id);
  }
  const selectedItem = selected ? [selected] : undefined;
  const onNewSession = (name: string) => newSession(name);
  const onClick = ({ key }: { key: string }) => selectSession(key);
  const items = sessions.map((session: Session) => (
    <Menu.Item key={session.id}>{session.name || session.id}</Menu.Item>
  ));
  return (
    <Layout.Sider
      className="sidebar"
      collapsible
      defaultCollapsed
      breakpoint="lg"
      collapsedWidth="0"
      theme="light"
    >
      <Spin spinning={loading}>
        <Menu
          className="menu"
          onClick={onClick}
          mode="inline"
          selectedKeys={selectedItem}
        >
          <Menu.ItemGroup title="Sessions">
            {items}
            <NewButton onValidate={onNewSession} />
          </Menu.ItemGroup>
        </Menu>
      </Spin>
    </Layout.Sider>
  );
};

export default connect(
  (state: AppState) => ({
    sessions: state.sessions.list,
    loading: state.sessions.loading,
    selected: state.sessions.selected
  }),
  (dispatch: Dispatch<Actions>) => ({
    fetch: () => dispatch(actions.fetchSessions.request()),
    selectSession: (sessionID: string) =>
      dispatch(actions.selectSession(sessionID)),
    newSession: (name: string) => dispatch(actions.newSession.request(name))
  })
)(SideBar);
