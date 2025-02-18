import React from "react";
import format from "../format/format";
import SetupCodeComponent from "../docs/setup_code";
import { invocation } from "../../proto/invocation_ts_proto";
import { TerminalComponent } from "../terminal/terminal";
import rpcService from "../service/rpc_service";
import { PauseCircle, PlayCircle } from "lucide-react";

interface Props {
  action: invocation.InvocationEvent;
  invocationId: string;
  dark: boolean;
}

interface State {
  stdErr: string;
  stdOut: string;
  cacheEnabled: boolean;
  loadingStdout: boolean;
  loadingStderr: boolean;
}

export default class ActionCardComponent extends React.Component {
  props: Props;

  state: State = {
    stdErr: "",
    stdOut: "",
    cacheEnabled: true,
    loadingStdout: true,
    loadingStderr: true,
  };

  componentDidMount() {
    console.log(this.props.action);
    this.fetchStdErr();
    this.fetchStdOut();
  }

  componentDidUpdate(prevProps: Props) {
    if (this.props.action !== prevProps.action) {
      this.fetchStdErr();
      this.fetchStdOut();
    }
  }

  fetchStdErr() {
    let logUrl = this.props.action?.buildEvent?.action?.stderr?.uri;

    if (!logUrl) {
      return;
    }

    if (!logUrl.startsWith("bytestream://")) {
      this.setState({ ...this.state, cacheEnabled: false });
      return;
    }

    this.setState({ ...this.state, loadingStderr: true });
    rpcService
      .fetchBytestreamFile(logUrl, this.props.invocationId)
      .then((contents: string) => {
        this.setState({
          ...this.state,
          stdErr: contents,
          loadingStderr: false,
        });
      })
      .catch(() => {
        this.setState({
          ...this.state,
          testLog: "Error loading bytestream stderr!",
        });
      });
  }

  fetchStdOut() {
    let logUrl = this.props.action?.buildEvent?.action?.stdout?.uri;

    if (!logUrl) {
      return;
    }

    if (!logUrl.startsWith("bytestream://")) {
      this.setState({ ...this.state, cacheEnabled: false });
      return;
    }

    this.setState({ ...this.state, loadingStdout: true });
    rpcService
      .fetchBytestreamFile(logUrl, this.props.invocationId)
      .then((contents: string) => {
        this.setState({
          ...this.state,
          stdout: contents,
          loadingStdout: false,
        });
      })
      .catch(() => {
        this.setState({
          ...this.state,
          testLog: "Error loading bytestream stdout!",
        });
      });
  }

  getStatusTitle(status: boolean) {
    return status ? "Passed" : "Failed";
  }

  render() {
    return (
      <div className="target-action-card">
        {this.props.action?.buildEvent?.action?.stderr?.uri && (
          <div
            className={`card ${this.state.cacheEnabled && (this.props.dark ? "dark" : "light-terminal")} ${
              this.props.action.buildEvent.action.success ? "card-success" : "card-failure"
            }`}>
            <PauseCircle className={`icon rotate-90 ${this.props.dark ? "white" : ""}`} />
            <div className="content">
              <div className="title">Error Log</div>
              <div className="test-subtitle">{this.getStatusTitle(this.props.action.buildEvent.action.success)}</div>
              {!this.state.cacheEnabled && (
                <div className="empty-state">
                  Log uploading isn't enabled for this invocation.
                  <br />
                  <br />
                  To enable action log uploading you must add GRPC remote caching. You can do so by checking{" "}
                  <b>Enable cache</b> below, updating your <b>.bazelrc</b> accordingly, and re-running your invocation:
                  <SetupCodeComponent />
                </div>
              )}
              {this.state.cacheEnabled && this.state.stdErr && (
                <div className="error-log">
                  <TerminalComponent value={this.state.stdErr} lightTheme={!this.props.dark} />
                </div>
              )}
              {this.state.cacheEnabled && this.state.loadingStderr && (
                <span>
                  <br />
                  Loading...
                </span>
              )}
              {this.state.cacheEnabled && !this.state.loadingStderr && !this.state.stdErr && (
                <span>
                  <br />
                  Empty log
                </span>
              )}
            </div>
          </div>
        )}

        {this.props.action?.buildEvent?.action?.stdout?.uri && (
          <div className={`card ${this.state.cacheEnabled && (this.props.dark ? "dark" : "light-terminal")}`}>
            <PauseCircle className={`rotate-90 ${this.props.dark ? "white" : ""}`} />
            <div className="content">
              <div className="title">Log</div>
              {!this.state.cacheEnabled && (
                <div className="empty-state">
                  Log uploading isn't enabled for this invocation.
                  <br />
                  <br />
                  To enable action log uploading you must add GRPC remote caching. You can do so by checking{" "}
                  <b>Enable cache</b> below, updating your <b>.bazelrc</b> accordingly, and re-running your invocation:
                  <SetupCodeComponent />
                </div>
              )}
              {this.state.cacheEnabled && this.state.stdOut && (
                <div className="error-log">
                  <TerminalComponent value={this.state.stdOut} lightTheme={!this.props.dark} />
                </div>
              )}
              {this.state.cacheEnabled && this.state.loadingStdout && (
                <span>
                  <br />
                  Loading...
                </span>
              )}
              {this.state.cacheEnabled && !this.state.loadingStdout && !this.state.stdOut && (
                <span>
                  <br />
                  Empty log
                </span>
              )}
            </div>
          </div>
        )}

        <div className={`card ${this.props.action.buildEvent.action.success ? "card-success" : "card-failure"}`}>
          <PlayCircle className="icon" />
          <div className="content">
            <div className="title">{this.props.action?.buildEvent?.action?.label}</div>
            <div className="test-subtitle">
              {this.props.action?.buildEvent?.action?.type} command exited with code{" "}
              {this.props.action?.buildEvent?.action?.exitCode}
            </div>
            <div>
              {this.props.action?.buildEvent?.action?.commandLine.map((commandLineArg) => (
                <div className="command-line-arg">{commandLineArg}</div>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }
}
