#!/bin/sh
if [ -n "$JAVA_HOME" ]; then
    JAVA="$JAVA_HOME"/bin/java
else
    JAVA=java
fi

echo "Using JAR_FILE = ${JAR_FILE}"

# Resolve the wildcard JAR_FILE pattern to the actual self-contained (uber) jar
USER_JAR=""
for f in ${JAR_FILE}; do
    USER_JAR="$f"
    break
done

if [ -z "$USER_JAR" ] || [ ! -f "$USER_JAR" ]; then
    echo "Unable to resolve a jar file from JAR_FILE=${JAR_FILE}" >&2
    exit 1
fi

echo "Running jar file ${USER_JAR}"

JAVA_OPTS="${JAVA_OPTS} -server"
JAVA_OPTS="${JAVA_OPTS} -Xmx1G -XX:+HeapDumpOnOutOfMemoryError"
JAVA_OPTS="${JAVA_OPTS} -XX:+UseG1GC -XX:+ParallelRefProcEnabled"
JAVA_OPTS="${JAVA_OPTS} -XX:MaxInlineLevel=20 -XX:MaxTrivialSize=12"
JAVA_OPTS="${JAVA_OPTS} --add-opens java.base/java.lang=ALL-UNNAMED"

# The jar is expected to be self-contained (all Gatling and simulation dependencies bundled in)
"$JAVA" $JAVA_OPTS -cp "$USER_JAR" io.gatling.app.Gatling "$@"
