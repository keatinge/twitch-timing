import React from 'react';
import ReactDOM from 'react-dom';
import Plot from "react-plotly.js"
import moment from "moment"
import './index.css';

function AllPlots(props) {

    if (props.resp === null || props.resp["IsError"]) {
        return null;
    }

    const plot_data = [
        {
            x: props.resp["BinTimes"],
            y: props.resp["BinFreqs"],
            line: {shape: "spline"},
            mode: "lines"
        }
    ];
    let n_ticks = 12;
    let x_tick_vals = [];
    for (let i = 0; i < n_ticks; i++) {
        x_tick_vals.push(i* 24/n_ticks);
    }
    const x_tick_str = x_tick_vals.map(x => moment().utc().hours(x).minutes(0).seconds(0).local().format("hh:mm a"));

    const plot_layout = {
        xaxis: {
            tickvals: x_tick_vals,
            ticktext: x_tick_str
        },
        autosize: true,
        // width: 960,
        // height: 540,
        title: `Streaming frequency for ${props.username}`
    };


    const dow_bar_data = [{
        x: ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"],
        y: props.resp.DowBins,
        type: "bar"
    }];


    const hist_data = [{
        x: props.resp.Durations,
        xbins: {
            size: 1,
        },
        type: "histogram"
    }];

    return (
        <div className="all-plots">
            <div className="overview">
                <h2 className="green">Successfully analyzed the last {props.resp.NumStreamsAnalyzed} streams from {props.resp.Username}</h2>
            </div>
            <div className="plot">
                <div className="info-box">
                    <h1 className="info-box-header ui header blue">Streaming times</h1>
                    <p className="white">What times does {props.resp.Username} usually stream?</p>
                </div>
                <div className="centered-plot">
                    <Plot data={plot_data} layout={plot_layout}/>
                </div>
            </div>
            <div className="plot">
                <div className="info-box">
                    <h1 className="info-box-header ui header blue">Streaming days</h1>
                    <p className="white">What day of the week does {props.resp.Username} usually stream?</p>
                </div>
                <div className="centered-plot">
                    <Plot data={dow_bar_data} layout={{}}/>
                </div>
            </div>
            <div className="plot">
                <div className="info-box">
                    <h1 className="info-box-header ui header blue">Stream duration</h1>
                    <p className="white">For how long does {props.resp.Username} usually stream?</p>
                </div>
                <div className="centered-plot">
                    <Plot data={hist_data} layout={{}}/>
                </div>
            </div>
        </div>
    )

}

class Root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            "username": "",
            "api_resp": null,
        }
    }

    show_error(err) {
        console.log(err)
    }

    start_lookup() {
        const to_lookup = this.state.username;
        const url = `${document.location.href}api/timings`
        fetch(url, {
            method: "POST",
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/json"
            },
            body: JSON.stringify({"Username": to_lookup})
        })
            .then(resp => resp.json())
            .then(json_resp => {
                this.setState({"api_resp": json_resp})
            })
            .catch(e => {
                this.show_error(e);
            });

        console.log(`Looking up ${to_lookup}`);
    }

    key_down(event) {
        if (event.key === "Enter") {
            this.start_lookup()
        }
    }

    handle_change(event) {
        this.setState({username: event.target.value})

    }

    render() {
        return (
            <div className="container">
                <div className="main-box">
                    <h1 className="big-header">Twitch Stream Timings</h1>
                    <div className="input-container">
                        <div className="ui massive action input">
                            <input type="text" placeholder="Username" onChange={(e) => this.handle_change(e)} onKeyDown={(e) => this.key_down(e)}/>
                            <button className="ui primary button" onClick={() => this.start_lookup()}>Analyze</button>
                        </div>
                    </div>
                </div>
                <AllPlots resp={this.state.api_resp} username={this.state.username}/>
            </div>
        )
    }

}

ReactDOM.render(
    <Root/>
    ,
    document.getElementById("root")
);