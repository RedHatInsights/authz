# handler package

The handler package inside the api folder implements server/router-independent handlers that can be called by whatever server/router one wants. Specific handler-wrappers may be defined in the server/router implementation, but call the same "internal" handler here.

This decouples API processing - e.g. checking request inputs for validity etc. - from the actual serverspecific implementation.