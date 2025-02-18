import React from "react";
import { PauseCircle } from "lucide-react";
import { TerminalComponent } from "../terminal/terminal";

interface Props {
  title: string;
  subtitle: string;
  contents: string;
  dark: boolean;
}

export default class TargetLogCardComponent extends React.Component {
  props: Props;

  render() {
    return (
      <div className={`card ${this.props.dark ? "dark" : "light-terminal"}`}>
        <PauseCircle className={`icon rotate-90 ${this.props.dark ? "white" : ""}`} />
        <div className="content">
          <div className="title">{this.props.title}</div>
          <div className="test-subtitle">{this.props.subtitle}</div>
          {this.props.contents && (
            <div className="test-log">
              <TerminalComponent value={this.props.contents} lightTheme={!this.props.dark} />
            </div>
          )}
        </div>
      </div>
    );
  }
}
