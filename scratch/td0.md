The standard TD(0) value update formula is:

$$V(S_t) \leftarrow V(S_t) + \alpha[R_{t+1} + \gamma V(S_{t+1}) - V(S_t)]$$

Where:
o  $V(S_t)$ is the estimated value of the current state.
o  $\alpha$ is the learning rate.
o  $R_{t+1}$ is the immediate reward received after transitioning.
o  $\gamma$ is the discount factor (determining the importance of future rewards).
o  The term $[R_{t+1} + \gamma V(S_{t+1}) - V(S_t)]$ is known as the TD Error ($\delta_t$), which represents the surprise or difference between expected and actual outcomes.
